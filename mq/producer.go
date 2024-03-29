package mq

import (
	"filestore-server/config"
	"github.com/streadway/amqp"
	"log"
)
var conn *amqp.Connection
var channel *amqp.Channel
// 如果异常关闭，会接收通知
var notifyClose chan *amqp.Error

func init() {
	// 是否开启异步转移功能，开启时才初始化rabbitMQ连接
	if !config.AsyncTransferEnable {
		return
	}
	if initChannel() {
		channel.NotifyClose(notifyClose)
	}
	// 断线自动重连
	go func() {
		for {
			select {
			case msg := <-notifyClose:
				conn = nil
				channel = nil
				log.Printf("onNotifyChannelClosed: %+v\n", msg)
				initChannel()
			}
		}
	}()
}
func initChannel()bool  {
	//判断channel是否创建过
	if channel != nil{
		return true
	}
	//获得rabbitmq一个连接
	conn, err := amqp.Dial(config.RabbitURL)
	if err != nil{
		log.Println(err.Error())
		return false
	}
	//打开channel ，用于消息的发布与接收
	channel, err = conn.Channel()
	if err != nil{
		log.Println(err.Error())
		return false
	}
	return true
}
//发布消息
func Publish(exchange, routingKey string,msg []byte)bool  {
	//判断channel是否正常
	if !initChannel() {
		return false
	}
	//执行消息发布动作
	err := channel.Publish(exchange,routingKey,
		false,false,
		amqp.Publishing{
			ContentType:     "text/plain",
			Body:            msg,
		})
	if err != nil{
		log.Println(err.Error())
		return false
	}
	return true
}