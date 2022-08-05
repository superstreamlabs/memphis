package server

import (
	"errors"
	"sort"
	"strings"
	"time"
)

func (s *Server) MemphisInitialized() bool {
	return s.GlobalAccount().JetStreamEnabled()
}

func (s *Server) MemphisAddStream(sc *StreamConfig) error {
	acc := s.GlobalAccount()
	_, err := acc.addStream(sc)

	return err
}

func (s *Server) MemphisAddConsumer(streamName string, cc *ConsumerConfig) error {
	acc := s.GlobalAccount()
	stream, err := acc.lookupStream(streamName)
	if err != nil {
		return err
	}

	_, err = stream.addConsumer(cc)

	return err
}

func (s *Server) MemphisGetConsumerInfo(streamName, consumerName string) (*ConsumerInfo, error) {
	acc := s.GlobalAccount()
	stream, err := acc.lookupStream(streamName)
	if err != nil {
		return nil, err
	}

	consumer := stream.lookupConsumer(consumerName)
	if consumer == nil {
		return nil, errors.New("Consumer doesn't exist")
	}

	return consumer.info(), nil
}

func (s *Server) MemphisRemoveStream(streamName string) error {
	acc := s.GlobalAccount()
	stream, err := acc.lookupStream(streamName)
	if err != nil {
		s.Errorf(err.Error())
		return err
	}

	_, err = stream.purge(nil)

	return err
}

func (s *Server) MemphisStreamInfo(streamName string) (*StreamInfo, error) {
	acc := s.GlobalAccount()
	mset, err := acc.lookupStream(streamName)
	if err != nil {
		return nil, err
	}
	config := mset.config()

	js, _ := s.getJetStreamCluster()

	info := StreamInfo{
		Created: mset.createdTime(),
		State:   mset.stateWithDetail(true),
		Config:  config,
		Domain:  s.getOpts().JetStreamDomain,
		Cluster: js.clusterInfo(mset.raftGroup()),
		Mirror:  mset.mirrorInfo(),
		Sources: mset.sourcesInfo(),
		// Alternates: js.streamAlternates(ci, config.Name),
	}

	return &info, nil
}

func (s *Server) MemphisAllStreamsInfo() []*StreamInfo {
	acc := s.GlobalAccount()
	streams := acc.streams()

	sort.Slice(streams, func(i, j int) bool {
		return strings.Compare(streams[i].cfg.Name, streams[j].cfg.Name) < 0
	})

	var res []*StreamInfo
	for _, mset := range streams {
		config := mset.config()
		res = append(res, &StreamInfo{
			Created: mset.createdTime(),
			State:   mset.state(),
			Config:  config,
			Domain:  s.getOpts().JetStreamDomain,
			Mirror:  mset.mirrorInfo(),
			Sources: mset.sourcesInfo(),
		})
	}

	return res
}

type MemphisHelperMsg struct {
	Subject string
	// Header    map[string]string
	Header    string
	Data      []byte
	Seq       uint64
	Timestamp time.Time
}

func (s *Server) MemphisGetMsgs(subjectName,
	streamName string,
	startSeq uint64,
	amount int,
	timeout time.Duration) ([]StoredMsg, error) {
	acc := s.GlobalAccount()
	stream, err := acc.lookupStream(streamName)
	if err != nil {
		return nil, err
	}

	var msgs []StoredMsg
	seq := startSeq

	timeoutCh := time.After(timeout)
	for i := 0; i < amount; i++ {
		select {
		case <-timeoutCh:
			return msgs, errors.New("MemphisGetMsgs timeout")
		default:
			pmsg := getJSPubMsgFromPool()
			sm, sseq, err := stream.store.LoadNextMsg(subjectName, false, seq, &pmsg.StoreMsg)
			seq = sseq + 1
			if sm == nil || err != nil {
				pmsg.returnToPool()
				return msgs, err
			}
			s.Noticef("buf: %v, msg:%v", string(sm.buf), string(sm.msg))
			msgs = append(msgs, StoredMsg{
				Subject:  sm.subj,
				Header:   sm.hdr,
				Data:     sm.buf,
				Sequence: sm.seq,
				Time:     time.Unix(0, sm.ts).UTC(),
			})
		}
	}

	return msgs, err
}

func (s *Server) MemphisGetSingleMsg(streamName string, msgSeq uint64) (*StoredMsg, error) {
	acc := s.GlobalAccount()
	stream, err := acc.lookupStream(streamName)
	if err != nil {
		return nil, err
	}
	return stream.getMsg(msgSeq)
}
