package dto

import "github.com/Mabarik667f/fsserver/internal/model/query"

func ToDocReadModelFromFullResponse(doc query.FullDocReadModel) DocReadModelResponse {
	return DocReadModelResponse{
		ID:        doc.ID,
		OwnerID:   doc.OwnerID,
		Name:      doc.Name,
		IsFile:    doc.IsFile,
		IsPublic:  doc.IsPublic,
		MimeType:  doc.MimeType,
		JSONData:  doc.JSONData,
		CreatedAt: doc.CreatedAt,
		Grants:    doc.Grants,
	}
}

func ToDocReadModelResponse(doc query.DocReadModel) DocReadModelResponse {
	return DocReadModelResponse{
		ID:        doc.ID,
		OwnerID:   doc.OwnerID,
		Name:      doc.Name,
		IsFile:    doc.IsFile,
		IsPublic:  doc.IsPublic,
		MimeType:  doc.MimeType,
		JSONData:  doc.JSONData,
		CreatedAt: doc.CreatedAt,
		Grants:    doc.Grants,
	}
}
