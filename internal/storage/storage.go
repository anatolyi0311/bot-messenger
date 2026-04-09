// Package storage - это пакет, который содержит логику взаимодействия с базой данных,
package storage

import (
	"database/sql"
	"fmt"

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

// Disable.
func (s *Storage) SaveIncomingID(chatID int64) (int, error) {
	var voiceID int
	// err := s.db.QueryRow(`
	// 	INSERT INTO voice_recognize (chat_id, process_step)
	//     VALUES ($1, $2)
	// 	RETURNING id
	// `, chatID, "REGISTER").Scan(&voiceID)
	return voiceID, nil
}

// SaveIncomingVoice - это метод, который сохраняет входящее голосовое сообщение в базе данных и
// возвращает его идентификатор (voiceID) для дальнейшей обработки.
func (s *Storage) SaveIncomingVoice(voiceBytes []byte, formatAudio string, chatID int64) (int, error) {
	var voiceID int
	// err := s.db.QueryRow(`
	// 	SELECT id
	// 	FROM voice_recognize
	// 	WHERE process_step = $1 AND chat_id = $2;
	// `, "REGISTER", chatID).Scan(&voiceID)
	// if _, err := s.db.Exec(`
	// 	UPDATE voice_recognize SET voice_data = $1, process_step = $2 WHERE id = $3;
	// `, voiceBytes, "BEGIN", voiceID); err != nil {
	// 	logrus.Info("storage.SaveIncomingVoice: update voice")
	// }
	err := s.db.QueryRow(`
			INSERT INTO voice_recognize (voice_data, format, chat_id, process_step) 
			VALUES ($1, $2, $3, $4)
			RETURNING id;
		`, voiceBytes, formatAudio, chatID, "BEGIN").Scan(&voiceID)
	logrus.Info("storage.SaveIncomingVoice: insert voice ID with voice data")
	return voiceID, err
}

// GetVoiceForUpload - это метод, который извлекает из базы данных список голосовых сообщений,
// которые находятся на этапе "BEGIN" и готовы для загрузки в Salute для распознавания речи.
func (s *Storage) GetVoiceForUpload() ([]models.UploadData, error) {
	// Выполнение SQL-запроса для получения голосовых сообщений, которые находятся на этапе "BEGIN"
	// и готовы для загрузки в Salute для распознавания речи.
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
	// Проход по результатам запроса и заполнение списка uploadDataList данными для загрузки в Salute
	// для распознавания речи.
	for rows.Next() {
		var uploadData models.UploadData
		// Проверка на наличие ошибок при сканировании результата и заполнении структуры uploadData
		// для каждого голосового сообщения.
		err := rows.Scan(&uploadData.VoiceID, &uploadData.ChatID, &uploadData.VoiceData)
		if err != nil {
			return nil, err
		}
		// Добавление данных голосового сообщения в список uploadDataList для загрузки в Salute
		// для распознавания речи.
		uploadDataList = append(uploadDataList, uploadData)
	}
	// Проверка на наличие ошибок при итерации по результатам запроса и заполнении списка uploadDataList.
	// Если список uploadDataList пустой, то возвращается ошибка с сообщением "no voices for upload"
	// для информирования вызывающего кода о том, что нет голосовых сообщений, готовых для загрузки
	// в Salute для распознавания речи.
	// Если список uploadDataList не пустой, то возвращается этот список для дальнейшей обработки.
	if err = rows.Err(); err != nil {
		return nil, err
	}
	// Проверка на наличие голосовых сообщений в списке uploadDataList и возвращение ошибки
	// с сообщением "no voices for upload", если список пустой, для информирования вызывающего кода о том,
	// что нет голосовых сообщений, готовых для загрузки в Salute для распознавания речи.
	if len(uploadDataList) < 1 {
		return nil, fmt.Errorf("no voices for upload") // errors.New("no voices for upload")
	}
	return uploadDataList, nil
}

// GetVoiceForRecognize - это метод, который извлекает из базы данных список голосовых сообщений,
// которые находятся на этапе "UPLOAD" и готовы для распознавания речи с помощью Salute.
func (s *Storage) GetVoiceForRecognize() ([]models.RecognizeData, error) {
	// Выполнение SQL-запроса для получения голосовых сообщений, которые находятся на этапе "UPLOAD"
	// и готовы для распознавания речи с помощью Salute.
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
	// Проверка на наличие ошибок при итерации по результатам запроса и заполнении списка recognizeDataList.
	// Если список recognizeDataList пустой, то возвращается ошибка с сообщением "no voices for recognize"
	// для информирования вызывающего кода о том, что нет голосовых сообщений, готовых для распознавания речи
	// с помощью Salute. Если список recognizeDataList не пустой, то возвращается этот список для дальнейшей обработки.
	// if len(recognizeDataList) < 1 {
	// 	return nil, fmt.Errorf("no voices for recognize") // errors.New("no voices for recognize")
	// }
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return recognizeDataList, nil
}

// GetVoiceForCheckStatus - это метод, который извлекает из базы данных список голосовых сообщений,
// которые находятся на этапе "RECOGNITION" и готовы для проверки статуса распознавания речи с помощью Salute.
func (s *Storage) GetVoiceForCheckStatus() ([]models.CheckStatusData, error) {
	// Выполнение SQL-запроса для получения голосовых сообщений, которые находятся на этапе "RECOGNITION"
	// и готовы для проверки статуса распознавания речи с помощью Salute.
	rows, err := s.db.Query(`
		SELECT id, chat_id, recognize_id
		FROM voice_recognize
		WHERE process_step = $1
	`, "RECOGNITION")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// Проход по результатам запроса и заполнение списка checkStatusList данными
	// для проверки статуса распознавания речи с помощью Salute.
	var checkStatusList []models.CheckStatusData
	for rows.Next() {
		var checkStatusData models.CheckStatusData
		err := rows.Scan(&checkStatusData.VoiceID, &checkStatusData.ChatID, &checkStatusData.RecognizeID)
		if err != nil {
			return nil, err
		}
		checkStatusList = append(checkStatusList, checkStatusData)
	}
	// Проверка на наличие ошибок при итерации по результатам запроса и заполнении списка checkStatusList.
	// Если при итерации возникает ошибка, то возвращается эта ошибка для дальнейшего анализа и устранения проблемы.
	// Если итерация прошла успешно, то возвращается список checkStatusList для дальнейшей обработки.
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return checkStatusList, nil
}

// GetVoiceForDownloadTranscription - это метод, который извлекает из базы данных список голосовых сообщений,
// которые находятся на этапе "WAIT" и готовы для загрузки транскрипции после распознавания речи с помощью Salute.
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
	// Проход по результатам запроса и заполнение списка downloadTranscriptionList данными
	// для загрузки транскрипции после распознавания речи с помощью Salute.
	var downloadTranscriptionList []models.DownloadTranscriptionData
	for rows.Next() {
		var downloadTranscription models.DownloadTranscriptionData
		err := rows.Scan(
			&downloadTranscription.VoiceID,
			&downloadTranscription.ChatID,
			&downloadTranscription.RespFileID,
		)
		if err != nil {
			return nil, err
		}
		// Добавление данных голосового сообщения в список downloadTranscriptionList
		// для загрузки транскрипции после распознавания речи с помощью Salute.
		// Если при сканировании данных возникает ошибка, то возвращается эта ошибка
		// для дальнейшего анализа и устранения проблемы.
		// Если данные были успешно отсканированы и добавлены в список downloadTranscriptionList,
		// то этот список будет возвращен для дальнейшей обработки.
		downloadTranscriptionList = append(downloadTranscriptionList, downloadTranscription)
	}
	// Проверка на наличие ошибок при итерации по результатам запроса и заполнении списка downloadTranscriptionList.
	// Если при итерации возникает ошибка, то возвращается эта ошибка для дальнейшего анализа и устранения проблемы.
	// Если итерация прошла успешно, то возвращается список downloadTranscriptionList для дальнейшей обработки.
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return downloadTranscriptionList, nil
}

// GetVoiceForCreateSummary - это метод, который извлекает из базы данных список голосовых сообщений,
// которые находятся на этапе "DOWNLOAD" и готовы для создания краткого содержания встречи.
func (s *Storage) GetVoiceForCreateSummary() ([]models.CreateSummaryData, error) {
	// Выполнение SQL-запроса для получения голосовых сообщений, которые находятся на этапе "DOWNLOAD"
	// и готовы для создания краткого содержания встречи.
	rows, err := s.db.Query(`
		SELECT id, chat_id, transcription
		FROM voice_recognize
		WHERE process_step = $1
	`, "DOWNLOAD")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Проход по результатам запроса и заполнение списка createSummaryList данными
	// для создания краткого содержания встречи.
	var createSummaryList []models.CreateSummaryData
	for rows.Next() {
		var createSummaryData models.CreateSummaryData
		err := rows.Scan(&createSummaryData.VoiceID, &createSummaryData.ChatID, &createSummaryData.Text)
		if err != nil {
			return nil, err
		}
		createSummaryList = append(createSummaryList, createSummaryData)
	}
	// Проверка на наличие ошибок при итерации по результатам запроса.
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return createSummaryList, nil
}

// GetSummaryByID - это метод, который получает краткое содержание по идентификатору встречи (voiceID)
// и идентификатору чата (chatID) из базы данных для предоставления пользователю.
func (s *Storage) GetSummaryByID(voiceID, chatID int64) (string, error) {
	// Выполнение SQL-запроса для получения краткого содержания по идентификатору встречи (voiceID)
	// и идентификатору чата (chatID) из базы данных для предоставления пользователю.
	var summary string
	err := s.db.QueryRow(`
		SELECT summary
		FROM voice_recognize
		WHERE id = $1 AND chat_id = $2
	`, voiceID, chatID).Scan(&summary)
	// Проверка на наличие ошибок при выполнении запроса и сканировании результата.
	if err != nil {
		logrus.Error("storage.GetSummaryByID: ", err, "summary:", summary, summary == "")
		return "", err
	}
	return summary, nil
}

// GetListSummaryID - это метод, который получает список идентификаторов встреч (IDs)
// по идентификатору чата (chatID) из базы данных для предоставления пользователю.
func (s *Storage) GetListSummaryID(chatID int64) ([]int64, error) {
	// Выполнение SQL-запроса для получения списка идентификаторов встреч (IDs)
	// по идентификатору чата (chatID)
	rows, err := s.db.Query(`
		SELECT id
		FROM voice_recognize
		WHERE chat_id = $1
		ORDER BY created_at	
	`, chatID)
	// Проверка на наличие ошибок при выполнении запроса.
	if err != nil {
		logrus.Error("storage.GetListSummaryID", err)
		return nil, err
	}
	defer rows.Close()

	// Проход по результатам запроса и заполнение среза IDs идентификаторами встреч
	// для предоставления пользователю.
	var IDs []int64
	for rows.Next() {
		var ID int64
		err := rows.Scan(&ID)
		// Проверка на наличие ошибок при сканировании результата.
		if err != nil {
			logrus.Error("Storage.GetListSummaryID", err)
			return nil, err
		}
		// Добавление идентификатора встречи в срез IDs для предоставления пользователю.
		IDs = append(IDs, ID)
	}
	// Проверка на наличие ошибок при итерации по результатам запроса.
	if err = rows.Err(); err != nil {
		logrus.Error("Storage.GetListSummaryID", err)
		return nil, err
	}
	// Проверка на наличие идентификаторов встреч в срезе IDs и возвращение ошибки sql.ErrNoRows,
	// если срез пустой.
	if len(IDs) == 0 {
		return IDs, sql.ErrNoRows
	}
	return IDs, nil
}

// FindByKey - это метод, который принимает идентификатор чата (chatID) и срез ключевых слов (keyWords),
// преобразует каждое ключевое слово в шаблон для поиска в базе данных
// и возвращает список идентификаторов встреч, которые соответствуют этим шаблонам.
func (s *Storage) FindByKey(chatID int64, keyWords []string) ([]int64, error) {
	// Выполнение SQL-запроса для получения списка идентификаторов встреч,
	// которые соответствуют шаблонам ключевых слов.
	rows, err := s.db.Query(`
		SELECT id
		FROM voice_recognize
		WHERE chat_id = $1 
			AND summary IS NOT NULL
			AND summary != ''
			AND summary ILIKE ANY($2)
		ORDER BY created_at	
	`, chatID, keyWords)
	// Проверка на наличие ошибок при выполнении запроса.
	if err != nil {
		logrus.Error("Storage.FindByKey", err)
		return nil, err
	}
	defer rows.Close()

	// Проход по результатам запроса и заполнение среза IDs идентификаторами встреч,
	// которые соответствуют шаблонам ключевых слов.
	var IDs []int64
	for rows.Next() {
		var ID int64
		err := rows.Scan(&ID)
		if err != nil {
			logrus.Error("Storage.FindByKey", err)
			return nil, err
		}
		// Добавление идентификатора встречи в срез IDs для предоставления пользователю.
		IDs = append(IDs, ID)
	}
	// Проверка на наличие ошибок при итерации по результатам запроса.
	if err = rows.Err(); err != nil {
		logrus.Error("Storage.FindByKey", err)
		return nil, err
	}
	// Проверка на наличие идентификаторов встреч в срезе IDs и возвращение ошибки sql.ErrNoRows,
	// если срез пустой.
	if len(IDs) == 0 {
		logrus.Error("Storage.FindByKey: empty select")
		return IDs, sql.ErrNoRows
	}
	return IDs, nil
}
