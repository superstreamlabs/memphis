package server

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
