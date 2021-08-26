package snowflake

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrWorkerIdOutOfRange = errors.New("snowflake: WorkerId out of range")
)

const (
	workerBits uint8 = 10
	numBits    uint8 = 12
	workerMax  int64 = ^(-1 << workerBits)
	numMax     int64 = ^(-1 << numBits)
	timeShift        = workerBits + numBits
	workShift        = numBits
	epoch      int64 = 1629973119
)

//var epoch int64

type Worker struct {
	mu          sync.Mutex
	timeMsStamp int64 // ms时间戳
	workerId    int64
	step        int64 //当前毫秒已经生成的id序列号
}

func NewWorker(workerId int64) (*Worker, error) {
	if workerId < 0 || workerId > workerMax {
		return nil, ErrWorkerIdOutOfRange
	}
	return &Worker{
		timeMsStamp: 0,
		workerId:    workerId,
		step:        0,
	}, nil
}
func (w *Worker) Now() int64 {
	return time.Now().Unix()
}

func (w *Worker) GetId() int64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	now := w.Now()
	if now == w.timeMsStamp {
		w.step++
		if w.step > numMax {
			w.step = 0
			for now < w.timeMsStamp {
				now = w.Now()
			}
			w.timeMsStamp = now
		}
	} else if now > w.timeMsStamp {
		w.step = 0
		w.timeMsStamp = now
	} else {
		//服务器发生了时钟回拨
		for now < w.timeMsStamp {
			now = w.Now()
		}
		w.step++
		if w.step > numMax {
			w.step = 0
			for now < w.timeMsStamp {
				now = w.Now()
			}
			w.timeMsStamp = now
		}
	}
	return (now-epoch)<<timeShift | (w.workerId << workShift) | w.step
}
