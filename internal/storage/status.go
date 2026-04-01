package storage

import "fmt"

// SetStatusRecognition - это метод, который обновляет статус обработки голосового сообщения в базе данных на "RECOGNITION" и сохраняет идентификатор распознавания (recognizeID) для заданного идентификатора голосового сообщения (voiceID).
func (s *Storage) SetStatusRecognition(voiceID int64, recognizeID string) error {
	_, err := s.db.Exec(`
		UPDATE voice_recognize 
		SET process_step = $2, recognize_id = $3
		WHERE id = $1
	`, voiceID, "RECOGNITION", recognizeID)
	return err
}

// SetStatusFail - это метод, который обновляет статус обработки голосового сообщения в базе данных на "FAIL" для заданного идентификатора голосового сообщения (voiceID).
func (s *Storage) SetStatusFail(voiceID int64) error {
	result, err := s.db.Exec(`
		UPDATE voice_recognize
		SET process_step = $2
		WHERE id = $1
	`, voiceID, "FAIL")

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("storage.SetStatusFail: rows affected = 0 for voice: %d", voiceID)
	}
	return nil
}

// SetStatusUpload - это метод, который обновляет статус обработки голосового сообщения в базе данных на "UPLOAD" и сохраняет идентификатор файла запроса (reqFileID) для заданного идентификатора голосового сообщения (voiceID).
func (s *Storage) SetStatusUpload(voiceID int64, reqFileID string) error {
	_, err := s.db.Exec(`
		UPDATE voice_recognize 
		SET process_step = $2, request_file_id = $3
		WHERE id = $1
	`, voiceID, "UPLOAD", reqFileID)
	return err
}
