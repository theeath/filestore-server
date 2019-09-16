package mq

import "log"
var done chan bool

//开始监听队列，获取消息
func StartConsume(qName,cName string,callback func(msg []byte)bool)  {
	//通过channel consumer获取信息通道
	msgs, err := channel.Consume(
		qName,cName,
		true,false,false,false,nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
	done = make(chan bool)
	go func() {
	for msgs := range msgs{
		//调用callback方法来处理新的消息
		processSuc := callback(msgs.Body)
		if !processSuc{
			//TODO:将任务写到另一个队列，用于异常情况的重试
		}

	}
	}()
	//done没有新的消息过来，则会一直阻塞
	<-done
	//关闭rabbitMQ
	channel.Close()
}