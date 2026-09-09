package entity

import (
	"github.com/Mabarik667f/fsserver/internal/model"
	"github.com/Mabarik667f/fsserver/internal/model/query"
)

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

func ToDocReadModel(doc DocRead) query.DocReadModel {
	return query.DocReadModel{
		ID:        doc.ID,
		OwnerID:   doc.OwnerID,
		Name:      doc.Name,
		IsFile:    doc.IsFile,
		IsPublic:  doc.IsPublic,
		MimeType:  doc.MimeType,
		JSONData:  doc.JSONData,
		FilePath:  doc.FilePath,
		CreatedAt: doc.CreatedAt,
		Grants:    doc.Grants,
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
