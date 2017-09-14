package qcache_inventory

/******************** Inventory Request
 Sends a query for a key or an IP and provides a back-channel, so that the requesting partner can block on the request
 until it arrives - honouring a timeout...
*/

import (
	"strings"
	"time"
	"github.com/docker/docker/api/types"
)

type ContainerRequest struct {
	IssuedAt 	time.Time
	Source 		string
	Timeout	 	time.Duration
	Name 		string
	ID 			string
	IP 			string
	Back 		chan Response
}

func NewContainerRequest(src string, to time.Duration) ContainerRequest {
	cr := ContainerRequest{
		IssuedAt: 	time.Now(),
		Source: 	src,
		Timeout:  	to,
		Back: 		make(chan Response, 5),
	}
	return cr
}


func NewIDContainerRequest(src, id string) ContainerRequest {
	cr := NewContainerRequest(src, time.Second)
	cr.ID = id
	return cr
}

func NewNameContainerRequest(src, name string) ContainerRequest {
	cr := NewContainerRequest(src, time.Second)
	cr.Name =  name
	return cr
}

func NewIPContainerRequest(src, ip string) ContainerRequest {
	cr := NewContainerRequest(src, time.Duration(2)*time.Second)
	cr.IP =  ip
	return cr
}

func (this ContainerRequest) Equal(other Response) bool {
	return this.EqualCnt(other.Container) && this.EqualIPS(other.Ips)
}

func (this ContainerRequest) EqualCnt(other *types.ContainerJSON) bool {
	idEqual := this.ID != "" && this.ID == other.ID
	nameEqual := this.Name != "" && this.Name == strings.Trim(other.Name, "/")
	return idEqual || nameEqual
}


func (this ContainerRequest) EqualIPS(ips []string) bool {
	if this.IP == "" {
		return true
	}
	for _, ip := range ips {
		if this.IP == ip {
			return true
		}
	}
	return false
}

func (cr *ContainerRequest) TimedOut() bool {
	tDiff := time.Now().Sub(cr.IssuedAt)
	return tDiff > cr.Timeout

}