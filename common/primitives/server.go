// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives

import (
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives/random"
)

type Server struct {
	ChainID interfaces.IHash
	Name    string
	Online  bool
	Replace interfaces.IHash
}

var _ interfaces.IServer = (*Server)(nil)
var _ interfaces.BinaryMarshallable = (*Server)(nil)

func (s *Server) Init() {
	if s.ChainID == nil {
		s.ChainID = NewZeroHash()
	}
	if s.Replace == nil {
		s.Replace = NewZeroHash()
	}
}

func RandomServer() interfaces.IServer {
	s := new(Server)
	s.Init()
	s.ChainID = RandomHash()
	s.Name = random.RandomString()
	s.Online = (random.RandInt()%2 == 0)
	s.Replace = RandomHash()
	return s
}

func (s *Server) IsSameAs(b interfaces.IServer) bool {
	serv := b.(*Server)
	if s.ChainID.IsSameAs(serv.ChainID) == false {
		return false
	}
	if s.Name != serv.Name {
		return false
	}
	if s.Online != serv.Online {
		return false
	}
	if s.Replace.IsSameAs(serv.Replace) == false {
		return false
	}
	return true
}

func (s *Server) MarshalBinary() ([]byte, error) {
	buf := new(Buffer)

	err := buf.PushBinaryMarshallable(s.ChainID)
	if err != nil {
		return nil, err
	}

	err = buf.PushString(s.Name)
	if err != nil {
		return nil, err
	}

	err = buf.PushBool(s.Online)
	if err != nil {
		return nil, err
	}

	err = buf.PushBinaryMarshallable(s.Replace)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (s *Server) UnmarshalBinaryData(p []byte) (newData []byte, err error) {
	s.Init()
	buf := NewBuffer(p)
	newData = p

	err = buf.PopBinaryMarshallable(s.ChainID)
	if err != nil {
		return
	}

	s.Name, err = buf.PopString()
	if err != nil {
		return
	}

	s.Online, err = buf.PopBool()
	if err != nil {
		return
	}

	err = buf.PopBinaryMarshallable(s.Replace)

	if err != nil {
		return
	}

	newData = buf.DeepCopyBytes()
	return
}

func (s *Server) UnmarshalBinary(p []byte) error {
	_, err := s.UnmarshalBinaryData(p)
	return err
}

func (s *Server) GetName() string {
	return s.Name
}

func (s *Server) GetChainID() interfaces.IHash {
	return s.ChainID
}

func (s *Server) String() string {
	return fmt.Sprintf("Server[:4]: %x", s.GetChainID().Bytes()[:10])
}

func (s *Server) IsOnline() bool {
	return s.Online
}

func (s *Server) SetOnline(o bool) {
	s.Online = o
}

func (s *Server) LeaderToReplace() interfaces.IHash {
	return s.Replace
}

func (s *Server) SetReplace(h interfaces.IHash) {
	s.Replace = h
}

func (e *Server) JSONByte() ([]byte, error) {
	return EncodeJSON(e)
}

func (e *Server) JSONString() (string, error) {
	return EncodeJSONString(e)
}