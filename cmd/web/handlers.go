package main

import (
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"snippetbox.lazarmrkic.com/internal/models"
	"snippetbox.lazarmrkic.com/internal/validator"
	"strconv"
)

// ovaj "struct" predstavlja podatke unutar forme i greške tokom validacije za njena polja
// sva njegova polja su u Pascal case, zato što moraju biti eksportovana - kako bi ih "html/template" paket pročitao prilikom renderovanja
type snippetCreateForm struct {
	// moramo da dodamo "struct" tagove, kako bi se vrijednosti iz HTML forme pravilno mapirale u različita polja struct-a
	Title   string `form:"title"`
	Content string `form:"content"`
	Expires int    `form:"expires"`
	// obrisaćemo "fieldErrors" polje
	// umjesto njega, "ugradićemo" Validator struct
	// to znači da će "snippetCreateForm" naslijediti sva polja i metode unutar njega
	validator.Validator `form:"-"`
}

type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

// "home" handler će postati metoda "application" struct-a:
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// novi "httprouter" tačno ubada "/" putanju, pa zbog toga uklanjamo "r.URL.Path != "/" provjeru
	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// pozivanje "newTemplateData()" helper metode, kako bi se dobio "templateData" struct
	// on sadrži "default" podatke (za sada, samo trenutnu godinu)
	// nakon toga, dodaje se "snippets" slice u njega
	data := app.newTemplateData(r)
	data.Snippets = snippets

	// prosljeđivanje podataka u "render()" helper funkciju
	app.render(w, r, http.StatusOK, "home.tmpl", data)
}

// "snippetView" handler će postati metoda "application" struct-a:
func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	// prilikom parsiranja request-ova u "httprouter", vrijednost svih imenovanih elemenataće biti sačuvana u "request" kontekstu
	// za sada je dovoljno da koristimo "ParamsFromContext()" funkciju da izvadimo "slice" koji sadrži nazive i vrijednosti ovih parametara
	params := httprouter.ParamsFromContext(r.Context())

	// koristićemo "params.byName()" metodu da izvadimo vrijednosti "id" imenovanog parametra iz gornjeg "slice"-a
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	snippet, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	// BITNO - naredni kod više nije potreban, pokriven je u "newTemplateData" metodi
	// "snippetView" handler treba da učita "flash" poruku (ukoliko ona postoji za trenutnog korisnika)
	// i nakon toga, da je proslijedi odgovarajućem HTML templejtu
	// pošto se ova poruka prikazuje samo jednom - potrebno ju je učitati i ukloniti nakon toga
	// zbog toga koristimo "PopString()" metodu
	//flash := app.sessionManager.PopString(r.Context(), "flash")
	// ukoliko želimo da ostavimo vrijednost unutar "session data", onda nam je dovoljna metoda "GetString()"

	data := app.newTemplateData(r)
	data.Snippet = snippet

	//data.Flash = flash

	app.render(w, r, http.StatusOK, "view.tmpl", data)
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	// prosljeđivanje "snippetCreateForm" instance u templejt
	// na ovaj način možemo da postavimo "default" vrijednost za formu, mimo "expires" polja
	data.Form = snippetCreateForm{
		Expires: 365,
	}

	app.render(w, r, http.StatusOK, "create.tmpl", data)
}

// "snippetCreate" handler će postati metoda "application" struct-a:
func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	// sad kad smo uveli novi "router", nema potrebe da provjeravamo da li je u pitanju POST request metoda

	// ograničenje za veličinu podataka unutar forme:
	//r.Body = http.MaxBytesReader(w, r.Body, 4096)

	// (STARI NAČIN) parsiranje forme iz "request"-a i provjera da li postoje neke greške
	//err := r.ParseForm()
	//if err != nil {
	//	app.clientError(w, http.StatusBadRequest)
	//	return
	//}

	// popunjavanje "snippetCreateForm" struct-a skupa sa "expires"/"title"/"content" poljima na sledeći način:
	var form snippetCreateForm
	err := app.decodePostForm(r, &form)
	//err = app.formDecoder.Decode(&form, r.PostForm)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// odličan blog post koji sadrži generalne metode za validaciju:
	// https://www.alexedwards.net/blog/validation-snippets-for-go#email-validation
	// validacija:
	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	form.CheckField(validator.PermittedValue(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7 or 365")

	// ukoliko neke od grešaka postoje, onda treba nanovo prikazati "create.tmpl" templejt
	// dinamički podaci će biti proslijeđeni u "Form" polje
	// takođe, treba poslati 402 HTTP status kod - koji pokazuje da je došlo do greške prilikom validacije
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "create.tmpl", data)
		return
	}

	// prosljeđivanje podataka ka bazi
	id, err := app.snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// preko "Put()" metode dodajemo string vrijednost i odgovarajući ključ ("flash")
	// "r.Context()" označava trenutni "request context"
	// gruba definicija - nešto gdje "session manager" PRIVREMENO čuva informacije, dok "handler"-i upravljaju zahtjevima
	app.sessionManager.Put(r.Context(), "flash", "Snippet successfully created!")

	// "redirect" putanja mora da se ažurira, kako bi se koristio novi, čistiji URL format
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupForm{}
	app.render(w, r, http.StatusOK, "signup.tmpl", data)
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "signup.tmpl", data)
		return
	}

	// dodavanje novog korisnika u bazu
	// ukoliko korisnik sa datim mejlom već postoji, onda treba prikazati "error message" na formi i ponovo je izrenderovati
	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldErrorKey("email", "Email address is already in use")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "signup.tmpl", data)
		} else {
			app.serverError(w, r, err)
		}
	}

	// ako je dodavanje prošlo OK, onda treba dodati "flash" poruku u sesiju, kako bi se potvrdilo da je "sign-up" prošao
	app.sessionManager.Put(r.Context(), "flash", "Your sign-up was successful. Please log in.")
	// nakon ovoga, vrši se redirekcija ka "login" stranici
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)

	return
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	app.render(w, r, http.StatusOK, "login.tmpl", data)
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var form userLoginForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "login.tmpl", data)
		return
	}

	// provjera da li su "user" kredencijali ispravni
	// ukoliko nisu - treba dodati "non-field error message" i ponovo prikazati "login" stranicu
	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "login.tmpl", data)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	// pozivom "RenewToken()" metode nad trenutnom sesijom se mijenja Session ID
	// uvijek treba generisati novi Session ID kada se kod nekog korisnika stanje autentifikacije ili privilegija promijeni
	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// u sesiju se dodaje ID-a trenutnog korisnika
	// nakon toga je "ulogovan"
	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)

	// preusmjeravanje korisnika na "createSnippet" stranicu
	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Logout the user...")
}
