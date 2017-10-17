package persistence

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/pborman/uuid"
	"gopkg.in/redis.v5"

	myredis "code-lib/redis"

	conf "subassembly/timer-controller/conf/timer"
)

const (
	defaultSessionIDFilePath = "${ProjectDir}/timerSession.txt"
	defaultRedisZsetKey      = "TimerMessageList:${SessionID}"
)

type Redis struct {
	cfg *conf.RedisPersistenceConf

	cli client

	id string

	redisZsetKey string
}

type client interface {
	redis.Cmdable
}

func CreateRedis(cfg *conf.RedisPersistenceConf) (persistence Redis, err error) {
	if cfg == nil {
		err = fmt.Errorf("cfg[conf.RedisPersistenceConf] is nil.")
		return
	}

	persistence = Redis{
		cfg: cfg,
	}

	// redis-cli
	err = persistence.newClient()
	if err != nil {
		err = fmt.Errorf("new redis client fail: %s", err)
		return
	}

	// sessionID
	err = persistence.newID()
	if err != nil {
		err = fmt.Errorf("new session id fail: %s", err)
		return
	}

	// redisZsetKey
	persistence.redisZsetKey = strings.Replace(defaultRedisZsetKey, "${SessionID}", persistence.id, -1)

	return
}

func NewRedis(cfg *conf.RedisPersistenceConf) (persistence *Redis, err error) {
	p, err := CreateRedis(cfg)
	return &p, err
}

// 存储数据
func (this *Redis) Set(key time.Time, data []byte) (err error) {
	nanosec := key.UnixNano()
	intCmd := this.cli.ZAdd(this.redisZsetKey, redis.Z{
		Score:  float64(nanosec),
		Member: data,
	})
	if err = intCmd.Err(); err != nil {
		return
	}

	return
}

// 获取少于当前时间的所有数据
func (this *Redis) Get(now time.Time) (datas [][]byte, err error) {
	nanosec := now.UnixNano()
	sliceCmd := this.cli.ZRangeByScore(this.redisZsetKey, redis.ZRangeBy{
		Min: strconv.FormatFloat(float64(0), 'f', -1, 64),
		Max: strconv.FormatFloat(float64(nanosec), 'f', -1, 64),
	})
	if err = sliceCmd.Err(); err != nil {
		return
	}

	datasstr := sliceCmd.Val()
	datas = make([][]byte, len(datasstr))
	for idx, str := range datasstr {
		datas[idx] = []byte(str)
	}

	return
}

// 删除时间范围内的所有数据, 范围: [start1, end1), [start2, end2) ...
func (this *Redis) Delete(start time.Time, end time.Time, pairs ...time.Time) (err error) {
	intCmd := this.cli.ZRemRangeByScore(this.redisZsetKey, // key
		strconv.FormatFloat(float64(start.UnixNano()), 'f', -1, 64), // min
		strconv.FormatFloat(float64(end.UnixNano()-1), 'f', -1, 64), // max = end - 1
	)
	if err = intCmd.Err(); err != nil {
		return err
	}

	if len(pairs)%2 == 0 {
		for i := 0; i < len(pairs); i++ {
			left := pairs[i]
			i++
			right := pairs[i]
			intCmd := this.cli.ZRemRangeByScore(this.redisZsetKey, // key
				strconv.FormatFloat(float64(left.UnixNano()+1), 'f', -1, 64),  // min preious end + 1
				strconv.FormatFloat(float64(right.UnixNano()-1), 'f', -1, 64), // max preious end - 1
			)
			if err = intCmd.Err(); err != nil {
				return
			}
		}
	}

	return
}

func (this *Redis) newClient() (err error) {
	// if have cluster, it will only connect to cluster, otherwise connect to single
	if this.cfg.Cluster != nil {
		clusterCli, nerr := myredis.NewRedisClusterClient(this.cfg.Cluster)
		if nerr != nil {
			err = nerr
			return
		}
		this.cli = clusterCli

	} else if this.cfg.Single != nil {
		singleCli, nerr := myredis.NewRedisClient(this.cfg.Single)
		if nerr != nil {
			err = nerr
			return
		}
		this.cli = singleCli

	} else {
		// where can I connect ?
		err = fmt.Errorf("have not redis node to connect.")
		return
	}

	return
}

func (this *Redis) newID() (err error) {
	idFile := this.cfg.SessionIDFile
	if idFile == "" {
		idFile = strings.Replace(defaultSessionIDFilePath, "${ProjectDir}", path.Dir(os.Args[0]), -1)
	}

	fmt.Println("idFile = ", idFile)

	_, err = os.Stat(idFile)
	if err != nil {
		if os.IsNotExist(err) {
			fp, nerr := os.OpenFile(idFile, os.O_WRONLY|os.O_CREATE, 0666)
			if nerr != nil {
				err = nerr
				return
			}
			defer fp.Close()

			this.id = uuid.New()
			_, err = io.Copy(fp, bytes.NewBufferString(this.id))
			if err != nil {
				return
			}

			return

		} else {
			return
		}
	}

	// file is exist
	fp, err := os.OpenFile(idFile, os.O_RDONLY, 0666)
	if err != nil {
		return
	}
	defer fp.Close()

	buffer := bytes.NewBuffer(make([]byte, 0))
	_, err = io.Copy(buffer, fp)
	if err != nil {
		return
	}

	this.id = buffer.String()

	return
}
