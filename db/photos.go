package db

type PhotosRow struct {
	Id         uint64 `db:"id"`
	Socket     uint64 `db:"socket"`
	User       uint64 `db:"user"`
	Added      int64  `db:"added"`
	Url        string `db:"url"`
	MediaId    string `db:"media_id"`
	Downloaded int    `db:"downloaded"`
}

func GetSocketPhotoStrings(socketId uint64) ([]string, error) {
	var rows []PhotosRow
	err := db.Select(&rows, `SELECT * FROM photos WHERE socket=?`, socketId)
	if err != nil {
		return nil, err
	}

	photos := make([]string, 0, len(rows))
	for _, row := range rows {
		mediaId := row.MediaId
		if mediaId == "" {
			mediaId = row.Url
		}
		photos = append(photos, mediaId)
	}
	return photos, nil
}

func GetSocketPhotoUrls(socketId uint64) ([]string, error) {
	var rows []PhotosRow
	err := db.Select(&rows, `SELECT * FROM photos WHERE socket=?`, socketId)
	if err != nil {
		return nil, err
	}

	photos := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.Url != "" {
			photos = append(photos, row.Url)
		}
	}
	return photos, nil
}

func GetAllPhotoRows() ([]PhotosRow, error) {
	var rows []PhotosRow
	err := db.Select(&rows, `SELECT * FROM photos`)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func SetPhotoUrl(photoId uint64, photoUrl string) error {
	_, err := db.Exec(`UPDATE photos SET url=? WHERE id=?`, photoUrl, photoId)
	return err
}
