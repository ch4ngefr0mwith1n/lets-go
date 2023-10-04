package main

import "snippetbox.lazarmrkic.com/internal/models"

// "templateData" će biti "struct" koji sadrži sve dinamičke podatke koje prosljeđujemo ka HTML templejtima
type templateData struct {
	Snippet models.Snippet
}
