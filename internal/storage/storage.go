// Package storage - это пакет, который содержит логику взаимодействия с базой данных,
package storage

import (
	"database/sql"
	"errors"

	"github.com/anatolyi0311/bot-messenger/internal/models"
	"github.com/sirupsen/logrus"
)

// Storage - это структура, которая будет использоваться для взаимодействия с базой данных.
type Storage struct {
	db *sql.DB
}

// New - это функция, которая инициализирует новый экземпляр Storage с заданной базой данных.
func New(db *sql.DB) *Storage {
	return &Storage{db: db}
}
func (s *Storage) SaveIncomingID(chatID int64) (int, error) {
	var voiceID int
	err := s.db.QueryRow(`
		INSERT INTO voice_recognize (chat_id, process_step) 
        VALUES ($1, $2)
		RETURNING id
	`, chatID, "REGISTER").Scan(&voiceID)
	return voiceID, err
}

// SaveIncomingVoice - это метод, который сохраняет входящее голосовое сообщение в базе данных и
// возвращает его идентификатор (voiceID) для дальнейшей обработки.
func (s *Storage) SaveIncomingVoice(voiceBytes []byte, chatID int64) (int, error) {
	var voiceID int

	err := s.db.QueryRow(`
		SELECT id
		FROM voice_recognize
		WHERE process_step = $1 AND chat_id = $2;
	`, "REGISTER", chatID).Scan(&voiceID)

	if _, err := s.db.Exec(`
		UPDATE voice_recognize SET voice_data = $1, process_step = $2 WHERE id = $3;
	`, voiceBytes, "BEGIN", voiceID); err != nil {
		err := s.db.QueryRow(`
			INSERT INTO voice_recognize (voice_data, chat_id, process_step) 
			VALUES ($1, $2, $3)
			RETURNING id;
		`, voiceBytes, chatID, "BEGIN").Scan(&voiceID)
		logrus.Info("storage.SaveIncomingVoice: insert voice ID with voice data")
		return voiceID, err
	}

	logrus.Info("storage.SaveIncomingVoice: update voice")
	return voiceID, err
}

// GetVoiceForUpload - это метод, который извлекает из базы данных список голосовых сообщений,
// которые находятся на этапе "BEGIN" и готовы для загрузки в Salute для распознавания речи.
func (s *Storage) GetVoiceForUpload() ([]models.UploadData, error) {
	// Выполнение SQL-запроса для получения голосовых сообщений, которые находятся на этапе "BEGIN" и готовы для загрузки в Salute для распознавания речи.
	rows, err := s.db.Query(`
		SELECT id, chat_id, voice_data
		FROM voice_recognize
		WHERE process_step = $1
	`, "BEGIN")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var uploadDataList []models.UploadData

	// Проход по результатам запроса и заполнение списка uploadDataList данными для загрузки в Salute для распознавания речи.
	for rows.Next() {
		var uploadData models.UploadData
		err := rows.Scan(&uploadData.VoiceID, &uploadData.ChatID, &uploadData.VoiceData)
		if err != nil {
			return nil, err
		}
		uploadDataList = append(uploadDataList, uploadData)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	if len(uploadDataList) < 1 {
		return nil, errors.New("no voices for upload")
	}
	return uploadDataList, nil
}

// GetVoiceForRecognize - это метод, который извлекает из базы данных список голосовых сообщений,
// которые находятся на этапе "UPLOAD" и готовы для распознавания речи с помощью Salute.
func (s *Storage) GetVoiceForRecognize() ([]models.RecognizeData, error) {
	// Выполнение SQL-запроса для получения голосовых сообщений, которые находятся на этапе "UPLOAD" и готовы для распознавания речи с помощью Salute.
	rows, err := s.db.Query(`
		SELECT id, chat_id, request_file_id
		FROM voice_recognize
		WHERE process_step = $1
	`, "UPLOAD")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recognizeDataList []models.RecognizeData

	// Проход по результатам запроса и заполнение списка recognizeDataList данными для распознавания речи.
	for rows.Next() {
		var recognizeData models.RecognizeData
		err := rows.Scan(&recognizeData.VoiceID, &recognizeData.ChatID, &recognizeData.ReqFileID)
		if err != nil {
			return nil, err
		}
		recognizeDataList = append(recognizeDataList, recognizeData)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return recognizeDataList, nil
}

func (s *Storage) GetVoiceForCheckStatus() ([]models.CheckStatusData, error) {
	rows, err := s.db.Query(`
		SELECT id, chat_id, recognize_id
		FROM voice_recognize
		WHERE process_step = $1
	`, "RECOGNITION")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checkStatusList []models.CheckStatusData

	for rows.Next() {
		var checkStatusData models.CheckStatusData
		err := rows.Scan(&checkStatusData.VoiceID, &checkStatusData.ChatID, &checkStatusData.RecognizeID)
		if err != nil {
			return nil, err
		}
		checkStatusList = append(checkStatusList, checkStatusData)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return checkStatusList, nil
}

func (s *Storage) GetVoiceForDownloadTranscription() ([]models.DownloadTranscriptionData, error) {
	rows, err := s.db.Query(`
		SELECT id, chat_id, response_file_id
		FROM voice_recognize
		WHERE process_step = $1
	`, "WAIT")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var downloadTranscriptionList []models.DownloadTranscriptionData

	for rows.Next() {
		var downloadTranscription models.DownloadTranscriptionData
		err := rows.Scan(&downloadTranscription.VoiceID, &downloadTranscription.ChatID, &downloadTranscription.RespFileID)
		if err != nil {
			return nil, err
		}
		downloadTranscriptionList = append(downloadTranscriptionList, downloadTranscription)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return downloadTranscriptionList, nil
}

func (s *Storage) GetVoiceForCreateSummary() ([]models.CreateSummaryData, error) {
	rows, err := s.db.Query(`
		SELECT id, chat_id, transcription
		FROM voice_recognize
		WHERE process_step = $1
	`, "DOWNLOAD")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var createSummaryList []models.CreateSummaryData

	for rows.Next() {
		var createSummaryData models.CreateSummaryData
		err := rows.Scan(&createSummaryData.VoiceID, &createSummaryData.ChatID, &createSummaryData.Text)
		if err != nil {
			return nil, err
		}
		createSummaryList = append(createSummaryList, createSummaryData)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return createSummaryList, nil
}

func (s *Storage) GetSummaryByID(voiceID, chatID int64) (string, error) {
	var summary string
	err := s.db.QueryRow(`
		SELECT summary
		FROM voice_recognize
		WHERE id = $1 AND chat_id = $2
	`, voiceID, chatID).Scan(&summary)
	if err != nil {
		return "", err
	}
	return summary, nil
}
