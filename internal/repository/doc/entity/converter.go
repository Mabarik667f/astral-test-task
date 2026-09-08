package entity

import "github.com/Mabarik667f/fsserver/internal/model"

func ToDoc(doc Doc) model.Doc {
	return model.Doc{
		ID:        doc.ID,
		OwnerID:   doc.OwnerID,
		Name:      doc.Name,
		IsFile:    doc.File,
		IsPublic:  doc.Public,
		MimeType:  doc.Mime,
		JSONData:  doc.JSON,
		FilePath:  doc.FilePath,
		CreatedAt: doc.Created,
	}
}

func ToDocDB(doc model.Doc) Doc {
	return Doc{
		ID:       doc.ID,
		OwnerID:  doc.OwnerID,
		Name:     doc.Name,
		File:     doc.IsFile,
		Public:   doc.IsPublic,
		Mime:     doc.MimeType,
		JSON:     doc.JSONData,
		FilePath: doc.FilePath,
		Created:  doc.CreatedAt,
	}
}
