package httpapi

type adapterConfigOption struct {
	Label string `json:"label"`
	Value any    `json:"value"`
}

type adapterConfigField struct {
	Key         string                `json:"key"`
	Label       string                `json:"label"`
	Type        string                `json:"type"`
	Placeholder string                `json:"placeholder,omitempty"`
	Grid        string                `json:"grid,omitempty"`
	Options     []adapterConfigOption `json:"options,omitempty"`
}

type adapterCatalogEntry struct {
	ResourceFields  []adapterConfigField `json:"resourceFields,omitempty"`
	OperationFields []adapterConfigField `json:"operationFields,omitempty"`
}

var adapterMetadataCatalog = map[string]adapterCatalogEntry{
	"supabase": {
		ResourceFields: []adapterConfigField{
			{Key: "table", Label: "Tabela", Type: "text", Placeholder: "vehicle_model_catalog"},
		},
		OperationFields: []adapterConfigField{
			{
				Key:   "method",
				Label: "Acao",
				Type:  "select",
				Options: []adapterConfigOption{
					{Label: "Select", Value: "select"},
					{Label: "Insert", Value: "insert"},
					{Label: "Update", Value: "update"},
					{Label: "Upsert", Value: "upsert"},
					{Label: "Delete", Value: "delete"},
				},
			},
		},
	},
	"mongo": {
		ResourceFields: []adapterConfigField{
			{Key: "collection", Label: "Colecao", Type: "text", Placeholder: "my_collection"},
		},
		OperationFields: []adapterConfigField{
			{
				Key:   "operation",
				Label: "Acao",
				Type:  "select",
				Options: []adapterConfigOption{
					{Label: "Find", Value: "find"},
					{Label: "Insert One", Value: "insertOne"},
					{Label: "Update One", Value: "updateOne"},
					{Label: "Delete One", Value: "deleteOne"},
				},
			},
		},
	},
	"firebase": {
		ResourceFields: []adapterConfigField{
			{Key: "collection", Label: "Colecao", Type: "text", Placeholder: "events"},
		},
		OperationFields: []adapterConfigField{
			{
				Key:   "operation",
				Label: "Acao",
				Type:  "select",
				Options: []adapterConfigOption{
					{Label: "Add", Value: "add"},
					{Label: "Set", Value: "set"},
				},
			},
			{Key: "docId", Label: "Document ID", Type: "text", Placeholder: "opcional para set"},
		},
	},
	"googledrive": {
		ResourceFields: []adapterConfigField{
			{Key: "folderId", Label: "Folder ID", Type: "text", Placeholder: "pasta de destino"},
		},
		OperationFields: []adapterConfigField{
			{Key: "fileName", Label: "Nome do arquivo", Type: "text", Placeholder: "export.json"},
		},
	},
	"googlespreadsheet": {
		ResourceFields: []adapterConfigField{
			{Key: "spreadsheetId", Label: "Spreadsheet ID", Type: "text", Placeholder: "planilha do Google Sheets"},
		},
		OperationFields: []adapterConfigField{
			{
				Key:   "action",
				Label: "Acao",
				Type:  "select",
				Options: []adapterConfigOption{
					{Label: "Append", Value: "append"},
					{Label: "Update", Value: "update"},
				},
			},
			{Key: "range", Label: "Range", Type: "text", Placeholder: "Sheet1!A1"},
		},
	},
	"mqtt": {
		ResourceFields: []adapterConfigField{
			{Key: "topic", Label: "Topico", Type: "text", Placeholder: "devices/telemetry"},
		},
		OperationFields: []adapterConfigField{
			{
				Key:   "qos",
				Label: "QoS",
				Type:  "select",
				Options: []adapterConfigOption{
					{Label: "0", Value: 0},
					{Label: "1", Value: 1},
					{Label: "2", Value: 2},
				},
			},
			{Key: "retained", Label: "Retained", Type: "boolean"},
		},
	},
	"jsonfile": {
		ResourceFields: []adapterConfigField{
			{Key: "fileName", Label: "Arquivo", Type: "text", Placeholder: "telemetry/output.ndjson"},
		},
		OperationFields: []adapterConfigField{
			{
				Key:   "format",
				Label: "Formato",
				Type:  "select",
				Options: []adapterConfigOption{
					{Label: "NDJSON", Value: "ndjson"},
					{Label: "JSON Array", Value: "json_array"},
				},
			},
			{
				Key:   "mode",
				Label: "Modo de escrita",
				Type:  "select",
				Options: []adapterConfigOption{
					{Label: "Append", Value: "append"},
					{Label: "Overwrite", Value: "overwrite"},
				},
			},
		},
	},
	"http": {
		ResourceFields: []adapterConfigField{
			{Key: "baseUrl", Label: "Base URL", Type: "text", Placeholder: "https://api.example.com"},
		},
		OperationFields: []adapterConfigField{
			{
				Key:   "method",
				Label: "Metodo",
				Type:  "select",
				Options: []adapterConfigOption{
					{Label: "GET", Value: "GET"},
					{Label: "POST", Value: "POST"},
					{Label: "PUT", Value: "PUT"},
					{Label: "PATCH", Value: "PATCH"},
					{Label: "DELETE", Value: "DELETE"},
				},
			},
			{Key: "path", Label: "Path", Type: "text", Placeholder: "/resource"},
		},
	},
	"no_auth": {
		ResourceFields: []adapterConfigField{
			{Key: "baseUrl", Label: "Base URL", Type: "text", Placeholder: "https://api.example.com"},
		},
		OperationFields: []adapterConfigField{
			{
				Key:   "method",
				Label: "Metodo",
				Type:  "select",
				Options: []adapterConfigOption{
					{Label: "GET", Value: "GET"},
					{Label: "POST", Value: "POST"},
					{Label: "PUT", Value: "PUT"},
					{Label: "PATCH", Value: "PATCH"},
					{Label: "DELETE", Value: "DELETE"},
				},
			},
			{Key: "path", Label: "Path", Type: "text", Placeholder: "/resource"},
		},
	},
	"api_key": {
		ResourceFields: []adapterConfigField{
			{Key: "baseUrl", Label: "Base URL", Type: "text", Placeholder: "https://api.example.com"},
		},
		OperationFields: []adapterConfigField{
			{
				Key:   "method",
				Label: "Metodo",
				Type:  "select",
				Options: []adapterConfigOption{
					{Label: "GET", Value: "GET"},
					{Label: "POST", Value: "POST"},
					{Label: "PUT", Value: "PUT"},
					{Label: "PATCH", Value: "PATCH"},
					{Label: "DELETE", Value: "DELETE"},
				},
			},
			{Key: "path", Label: "Path", Type: "text", Placeholder: "/resource"},
		},
	},
	"basic_auth": {
		ResourceFields: []adapterConfigField{
			{Key: "baseUrl", Label: "Base URL", Type: "text", Placeholder: "https://api.example.com"},
		},
		OperationFields: []adapterConfigField{
			{
				Key:   "method",
				Label: "Metodo",
				Type:  "select",
				Options: []adapterConfigOption{
					{Label: "GET", Value: "GET"},
					{Label: "POST", Value: "POST"},
					{Label: "PUT", Value: "PUT"},
					{Label: "PATCH", Value: "PATCH"},
					{Label: "DELETE", Value: "DELETE"},
				},
			},
			{Key: "path", Label: "Path", Type: "text", Placeholder: "/resource"},
		},
	},
	"bearer_token": {
		ResourceFields: []adapterConfigField{
			{Key: "baseUrl", Label: "Base URL", Type: "text", Placeholder: "https://api.example.com"},
		},
		OperationFields: []adapterConfigField{
			{
				Key:   "method",
				Label: "Metodo",
				Type:  "select",
				Options: []adapterConfigOption{
					{Label: "GET", Value: "GET"},
					{Label: "POST", Value: "POST"},
					{Label: "PUT", Value: "PUT"},
					{Label: "PATCH", Value: "PATCH"},
					{Label: "DELETE", Value: "DELETE"},
				},
			},
			{Key: "path", Label: "Path", Type: "text", Placeholder: "/resource"},
		},
	},
}

func adapterMetadataFor(authTypeID string, serverTypeID string) adapterCatalogEntry {
	if entry, ok := adapterMetadataCatalog[authTypeID]; ok {
		return entry
	}
	if entry, ok := adapterMetadataCatalog[serverTypeID]; ok {
		return entry
	}
	return adapterCatalogEntry{}
}
