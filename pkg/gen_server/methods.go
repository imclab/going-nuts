// "Erlang OTP gen_server"-like Concurrency Pattern for Go
package gen_server

import (
	"log"
	"runtime"
)

// loop implements the internal GenServer goroutine loop
func (self *GenServer) loop() {
  for ; ; {
		self.status = READY
		select {
		case msg := <- self.ch :
			self.status = BUSY
			switch cmsg := msg.(type) {
			case CallMessage:
				self.handle_call(&cmsg)
			case CastMessage:
				self.handle_cast(&cmsg)
			}
		case cmd := <- self.control_ch :
			self.status = BUSY
			switch ccmd := cmd.(type) {
			case initControlMessage:
				self.handle_init(&ccmd)
			case stopControlMessage:
				self.handle_stop(&ccmd)
			}
			}
	}
}

// log will print log messages if self.debug is true
func (self *GenServer) log(log_msg string, log_data interface {}) {
	if self.debug {
		log.Print(log_msg,log_data)
	}
}

// SetDebug enable/disable debugging messages
func (self *GenServer) SetDebug(debug bool) {
	self.debug = debug
}

// handle_init will be called to handle incoming InitControlMessages
func (self *GenServer) handle_init(cmd *initControlMessage) {
  self.log("RECEIVED INIT ",cmd)
	self.status = READY
	cmd.ReplyChannel <- ReplyMessage{Ok: true}
}

// handle_stop will be called to handle incoming StopControlMessages
func (self *GenServer) handle_stop(cmd *stopControlMessage) {
	defer func() {
		// TODO: error handling
		self.status = STOPPED
		cmd.ReplyChannel <- ReplyMessage{Ok: true}
	}()
	self.log("RECEIVED STOP ",cmd)
	runtime.Goexit()
}

// handle_cast will be called to handle incoming CastMessages
func (self *GenServer) handle_cast(cast *CastMessage) {
	self.log("RECEIVED CAST ",cast)
	self.impl.HandleCast(cast)
}

// handle_call will be called to handle incoming CallMessages
func (self *GenServer) handle_call(call *CallMessage) {
	self.log("RECEIVED CALL ",call)
	self.impl.HandleCall(call)
}

// GetStatus returns the current GenServer Status
func (self *GenServer) GetStatus() int {
	return self.status
}


func (self *GenServer) Start(impl IGenServerImpl) {
	self.status = STARTING;
	ch := make(MessageChannel)
	control_ch := make(controlChannel)
	self.ch = ch
	self.control_ch = control_ch
	self.impl = impl
	go self.loop()  
	reply_ch := make(ReplyMessageChannel)
	self.control_ch <- initControlMessage{ReplyChannel: reply_ch}
	<- reply_ch
}

func (self *GenServer) Stop() ReplyMessage {
	reply_ch := make(ReplyMessageChannel)
	self.control_ch <- stopControlMessage{ReplyChannel: reply_ch}
	return <- reply_ch
}


func (self *GenServer) Cast(name string, args Data) {
	self.ch <- CastMessage{Name: name, Args: args}
}

func (self *GenServer) Call(name string, args Data) ReplyMessage {
	reply_ch := make(ReplyMessageChannel)
	self.ch <- CallMessage{Name: name, Args: args, ReplyChannel: reply_ch}
	return <- reply_ch
}

func (self *CallMessage) Reply(ok bool, result Data) {
	self.ReplyChannel <- ReplyMessage{Ok: ok, Result: result}
}

