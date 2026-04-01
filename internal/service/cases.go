package service

func (w *Worker) SaveIncomingVoice(voiceBytes []byte, chatID int64) (int, error) {
	return w.storage.SaveIncomingVoice(voiceBytes, chatID)
}
