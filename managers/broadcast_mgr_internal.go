package managers

import "github.com/keydotcat/server/models"

type ibmRegister struct {
	sid  string
	resp chan (<-chan *Broadcast)
}

type InternalBroadcasterMgr struct {
	regChan    chan ibmRegister
	delChan    chan string
	sourceChan chan *Broadcast
	stopChan   chan bool
	clients    map[string]chan<- *Broadcast
}

func NewInternalBroadcasterMgr() BroadcasterMgr {
	ibm := &InternalBroadcasterMgr{
		make(chan ibmRegister),
		make(chan string),
		make(chan *Broadcast, 5), //To prevent locks in high througthputh moments
		make(chan bool),
		make(map[string]chan<- *Broadcast),
	}
	go ibm.run()
	return ibm
}

func (ibm *InternalBroadcasterMgr) del(sid string) {
	close(ibm.clients[sid])
	delete(ibm.clients, sid)
}

func (ibm *InternalBroadcasterMgr) send(sid string, c chan<- *Broadcast, b *Broadcast) {
	for {
		if _, ok := ibm.clients[sid]; !ok {
			break
		}
		select {
		case c <- b:
			break
		case sid := <-ibm.delChan:
			ibm.del(sid)
		}
	}
}

func (ibm *InternalBroadcasterMgr) run() {
	var b *Broadcast
	for {
		select {
		case <-ibm.stopChan:
			for _, c := range ibm.clients {
				close(c)
			}
			break
		case b = <-ibm.sourceChan:
			for sid, c := range ibm.clients {
				ibm.send(sid, c, b)
			}
		case r := <-ibm.regChan:
			bc := make(chan *Broadcast)
			ibm.clients[r.sid] = bc
			r.resp <- bc
		case sid := <-ibm.delChan:
			ibm.del(sid)
		}
	}
}

func (ibm *InternalBroadcasterMgr) Subscribe(sid string) <-chan *Broadcast {
	r := ibmRegister{sid, make(chan (<-chan *Broadcast))}
	ibm.regChan <- r
	return <-r.resp
}

func (ibm *InternalBroadcasterMgr) Unsubscribe(sid string) {
	ibm.delChan <- sid
}

func (ibm *InternalBroadcasterMgr) Send(team, vault, action string, secret *models.Secret) {
	ibm.sourceChan <- createBroadcast(team, vault, action, secret)
}

func (ibm *InternalBroadcasterMgr) Stop() {
	ibm.stopChan <- true
}
