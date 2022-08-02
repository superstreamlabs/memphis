package server

import (
	"errors"
	"sort"
	"strings"
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
