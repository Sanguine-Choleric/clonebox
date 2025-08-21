package main

import (
	"crypto/md5"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"snippetbox/internal/models"
	"snippetbox/internal/validator"
	"strings"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"google.golang.org/genai"
)

type snippetCreateForm struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
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

type accountPasswordUpdateForm struct {
	CurrentPassword     string `form:"current_password"`
	NewPassword         string `form:"new_password"`
	NewPasswordConfirm  string `form:"new_password_confirm"`
	validator.Validator `form:"-"`
}

type linkShortenForm struct {
	OriginalLink        string `form:"original_link"`
	ShortLink           string `form:"short_link"`
	validator.Validator `form:"-"`
}

type fileShareForm struct {
	Filename string `form:"file_name"`
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	links, err := app.links.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Snippets = snippets
	data.Links = links

	app.render(w, http.StatusOK, "home.tmpl.html", data)
}

func (app *application) about(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, http.StatusOK, "about.tmpl.html", data)
}

func (app *application) tools(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, http.StatusOK, "tools.tmpl.html", data)
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	public_id := params.ByName("public_id")

	snippet, err := app.snippets.Get(public_id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	data := app.newTemplateData(r)
	data.Snippet = snippet

	app.render(w, http.StatusOK, "view.tmpl.html", data)
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	data.Form = snippetCreateForm{
		Expires: 365,
	}
	app.render(w, http.StatusOK, "create.tmpl.html", data)
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	var form snippetCreateForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	form.CheckField(validator.PermittedValue(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7 or 365")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "create.tmpl.html", data)
		return
	}

	public_id, err := app.snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Snippet successfully saved")

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%s", public_id), http.StatusSeeOther)
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupForm{}

	app.render(w, http.StatusOK, "signup.tmpl.html", data)
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validates the form contents using helper functions.
	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "signup.tmpl.html", data)
		return
	}

	// Try to create a new user record in the database. If the email already exists then add an error message to the form and re-display it.
	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email is already in use")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "signup.tmpl.html", data)
		} else {
			app.serverError(w, err)
		}

		return
	}

	// Otherwise add a confirmation flash message to the session confirming that signup worked.
	app.sessionManager.Put(r.Context(), "flash", "Signup success")

	// Redirect the user to the login page.
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}

	app.render(w, http.StatusOK, "login.tmpl.html", data)
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var form userLoginForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.serverError(w, err)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "login.tmpl.html", data)
		return
	}

	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or Password is incorrect")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnauthorized, "login.tmpl.html", data)
		} else {
			app.serverError(w, err)
		}
		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "authenticatedUserId", id)

	path := app.sessionManager.GetString(r.Context(), "originalPath")
	if path != "" {
		app.sessionManager.Remove(r.Context(), "originalPath")
		http.Redirect(w, r, path, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/about", http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	// Uses the RenewToken() method on the current session to change the session ID
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUserId")

	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) accountView(w http.ResponseWriter, r *http.Request) {
	userId := app.sessionManager.GetInt(r.Context(), "authenticatedUserId")
	user, err := app.users.Get(userId)
	if errors.Is(err, models.ErrNoRecord) {
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
	} else if err != nil {
		app.serverError(w, err)
	}

	data := app.newTemplateData(r)
	data.User = user
	app.render(w, http.StatusOK, "account.tmpl.html", data)
	//fmt.Fprintf(w, "%+v", user)
}

func (app *application) accountPasswordUpdate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = accountPasswordUpdateForm{}
	app.render(w, http.StatusOK, "change_password.tmpl.html", data)
}

func (app *application) accountPasswordUpdatePost(w http.ResponseWriter, r *http.Request) {
	var form accountPasswordUpdateForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.serverError(w, err)
		return
	}

	form.CheckField(validator.NotBlank(form.CurrentPassword), "currentPassword", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.NewPassword), "newPassword", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.NewPassword, 8), "newPassword", "This field must be at least 8 characters long")
	form.CheckField(validator.NotBlank(form.NewPasswordConfirm), "newPasswordConfirm", "This field cannot be blank")
	form.CheckField(form.NewPassword == form.NewPasswordConfirm, "newPasswordConfirm", "Passwords do not match")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "change_password.tmpl.html", data)
		return
	}

	userId := app.sessionManager.GetInt(r.Context(), "authenticatedUserId")
	err = app.users.PasswordUpdate(userId, form.CurrentPassword, form.NewPassword)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddFieldError("currentPassword", "Entered password is not your current password")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "change_password.tmpl.html", data)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Password successfully changed")
	http.Redirect(w, r, "/account/view", http.StatusSeeOther)

}

func (app *application) linkShorten(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = linkShortenForm{}

	app.render(w, http.StatusOK, "link_shorten.tmpl.html", data)
}

func (app *application) linkShortenPost(w http.ResponseWriter, r *http.Request) {
	var form linkShortenForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Valid URL checks
	originalLink := form.OriginalLink
	if !strings.HasPrefix(originalLink, "http://") && !strings.HasPrefix(originalLink, "https://") {
		originalLink = "https://" + originalLink
	}
	form.CheckField(validator.IsURL(originalLink), "originalLink", "This field must be a valid URL")
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "link_shorten.tmpl.html", data)
		return
	}

	// Check if user entered link is already in db. If so, directly render that
	exists, err := app.links.Exists(originalLink)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			app.serverError(w, err)
			return
		}
		app.errorLog.Printf("%v", err)
	}

	if exists {
		short, err := app.links.GetShort(originalLink)
		if err != nil {
			app.serverError(w, err)
		}

		data := app.newTemplateData(r)
		form.ShortLink = fmt.Sprintf("%s/shorten/%s", r.Host, short)
		data.Form = form
		app.render(w, http.StatusOK, "link_shorten.tmpl.html", data)
		return
	}

	// Begin shortening logic
	hash := sha256.Sum256([]byte(originalLink))
	shortLink := fmt.Sprintf("%x", hash[0:3]) // Does this result in collisions?

	// TODO Definitely unit test this
	// If originalLink is a duplicate, earlier code would have caught it (and served existing)
	// If shortLink is a duplicate, re-hash until no hash collision
	err = app.links.Insert(originalLink, shortLink)
	for errors.Is(err, models.ErrDuplicateLink) {
		betterHash := sha256.Sum256([]byte(shortLink + originalLink))
		shortLink = fmt.Sprintf("%x", betterHash[0:3])

		err = app.links.Insert(originalLink, shortLink)
	}

	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	form.OriginalLink = originalLink
	form.ShortLink = fmt.Sprintf("%s/shorten/%s", r.Host, shortLink)
	data.Form = form
	app.render(w, http.StatusOK, "link_shorten.tmpl.html", data)
}

func (app *application) linkRedirect(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	hash := params.ByName("hash")

	originalLink, err := app.links.GetOriginal(hash)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.clientError(w, http.StatusNotFound)
			return
		}
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, originalLink, http.StatusSeeOther)
}

func (app *application) fileUpload(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = fileShareForm{}

	app.render(w, http.StatusOK, "file_upload.tmpl.html", data)
}

func (app *application) fileUploadPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(100 << 20) // 100MB -> 2^20 is 1MB
	if err != nil {
		app.serverError(w, err)
		return
	}

	uploadedFile, header, err := r.FormFile("file")
	if err != nil {
		app.serverError(w, err)
		return
	}
	defer uploadedFile.Close()

	// Unique file name - store in db?
	fileUUID := uuid.New().String()
	storagePath := "/clonebox/uploads/"
	dst, err := os.Create(storagePath + fileUUID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	defer dst.Close()

	// Need to use teeReader to avoid weird read-once limitation on multipart form file
	// This handles saving to disk and checksum calc in same pass
	h := md5.New()
	teeReader := io.TeeReader(uploadedFile, h)
	_, err = io.Copy(dst, teeReader)
	if err != nil {
		app.serverError(w, err)
		return
	}
	checksum := fmt.Sprintf("%x", h.Sum(nil))

	// Begin DB Storage
	// Checking for existing file w/ matching checksum. If exists, delete previously copied. Using this approach because of
	// io.Reader limitations -- TODO: Find a better way
	existingFile, err := app.files.GetByChecksum(checksum)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			// Currently storing storagePath in case i use a per-user dir approach
			err = app.files.Insert(header.Filename, fileUUID, int(header.Size), checksum, storagePath)
			if err != nil {
				app.serverError(w, err)
				return
			}

			http.Redirect(w, r, fmt.Sprintf("/file/view/%s", fileUUID), http.StatusSeeOther)
			return
		}
		app.serverError(w, err)
		return
	}

	// Delete dupe upload
	err = os.Remove(storagePath + fileUUID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/file/view/%s", existingFile.FileUUID), http.StatusSeeOther)
}

func (app *application) fileView(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	fileNameUUID := params.ByName("uuid")

	uploadedFile, err := app.files.GetByUUID(fileNameUUID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.clientError(w, http.StatusNotFound)
			return
		}
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.File = uploadedFile

	app.render(w, http.StatusOK, "file_download.tmpl.html", data)
}

func (app *application) fileDownload(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	fileNameUUID := params.ByName("uuid")

	uploadedFile, err := app.files.GetByUUID(fileNameUUID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.clientError(w, http.StatusNotFound)
			return
		}
		app.serverError(w, err)
		return
	}

	if _, err = os.Stat(uploadedFile.StoragePath); err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	fileNameAndPath := uploadedFile.StoragePath + uploadedFile.FileUUID
	http.ServeFile(w, r, fileNameAndPath)
}

func (app *application) billSplit(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	//data.Form = billSplitForm{}

	app.render(w, http.StatusOK, "bill_split.tmpl.html", data)
}

func (app *application) billSplitPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(5 << 20)
	if err != nil {
		app.serverError(w, err)
		return
	}

	receiptImage, _, err := r.FormFile("file")
	if err != nil {
		app.serverError(w, err)
		return
	}
	defer receiptImage.Close()

	bytes, err := io.ReadAll(receiptImage)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// TODO Use a flash error instead
	if http.DetectContentType(bytes) != "image/jpeg" {
		app.clientError(w, http.StatusUnsupportedMediaType)
		return
	}

	// Test response
	var result *genai.GenerateContentResponse
	if *app.debug {
		testString := `[
  {
    "name": "Moscow Mule",
    "price": 8.9,
    "quantity": 3
  },
  {
    "name": "Franziskaner Hefeweizen",
    "price": 5.5,
    "quantity": 4
  },
  {
    "name": "Riq. Sauvignon Blanc",
    "price": 6.9,
    "quantity": 1
  },
  {
    "name": "Krombacher Pils 0,4L",
    "price": 5.5,
    "quantity": 1
  },
  {
    "name": "K1. House Salad",
    "price": 5.9,
    "quantity": 1
  },
  {
    "name": "Starter Caesar Salad",
    "price": 6.9,
    "quantity": 1
  },
  {
    "name": "Caesar Garnelen",
    "price": 17.9,
    "quantity": 1
  },
  {
    "name": "Starter Caesar Salad",
    "price": 6.9,
    "quantity": 1
  },
  {
    "name": "240g Hähnchenbrustfilet",
    "price": 22.9,
    "quantity": 1
  },
  {
    "name": "Süßkartoffel Fries",
    "price": 5.9,
    "quantity": 1
  },
  {
    "name": "M 300g Arg. Rumpsteak",
    "price": 34.9,
    "quantity": 1
  },
  {
    "name": "M 300g Arg. Huftsteak",
    "price": 34.9,
    "quantity": 1
  },
  {
    "name": "M 400g Arg. Entrecôte",
    "price": 49.9,
    "quantity": 1
  },
  {
    "name": "Baked Kartöffeli",
    "price": 5.9,
    "quantity": 1
  },
  {
    "name": "M 300g Arg. Huftsteak",
    "price": 34.9,
    "quantity": 1
  },
  {
    "name": "Knoblauch-Kräuterbutter",
    "price": 3.9,
    "quantity": 1
  },
  {
    "name": "Trüffel Butter",
    "price": 3.9,
    "quantity": 1
  },
  {
    "name": "Aqua Panna 0,75L",
    "price": 6.9,
    "quantity": 1
  },
  {
    "name": "Whiskey Sour",
    "price": 8.9,
    "quantity": 1
  },
  {
    "name": "Oreo Brownie",
    "price": 8.9,
    "quantity": 1
  },
  {
    "name": "Latte Macchiato",
    "price": 3.9,
    "quantity": 1
  },
  {
    "name": "Apple Pie",
    "price": 8.9,
    "quantity": 2
  },
  {
    "name": "Jägermeister 4cl",
    "price": 4.9,
    "quantity": 4
  }
]`
		result = &genai.GenerateContentResponse{
			Candidates: []*genai.Candidate{
				&genai.Candidate{
					Content: &genai.Content{
						Parts: []*genai.Part{
							&genai.Part{Text: testString},
						},
					},
				},
			},
		}
	} else {
		//Actual LLM call
		parts := []*genai.Part{
			genai.NewPartFromBytes(bytes, "image/jpeg"),
			genai.NewPartFromText(
				"Give me the name, price, and quantity of each item in this receipt. Include tax and other fees as its own item with a name (Tax + fees), price, quantity (1). If the image isn't a receipt, only give a single item with a name (Error), price (0), quantity (0)",
			),
		}

		contents := []*genai.Content{
			genai.NewContentFromParts(parts, genai.RoleUser),
		}

		result, err = app.llmClient.Models.GenerateContent(
			r.Context(),
			"gemini-2.0-flash-lite",
			contents,
			app.llmConfig,
		)

		if err != nil {
			app.serverError(w, fmt.Errorf("failed llm generate: %w", err))
			return
		}
	}

	var items []models.BillItem
	err = json.Unmarshal([]byte(result.Text()), &items)
	if err != nil {
		app.serverError(w, fmt.Errorf("failed llm unmarshal: %w", err))
		return
	}

	app.infoLog.Println(result.Text())

	data := app.newTemplateData(r)
	data.BillItems = items
	app.render(w, http.StatusOK, "bill_split.tmpl.html", data)
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
