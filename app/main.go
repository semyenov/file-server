package main

import (
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const tpl1 = `
<html>
<head>
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>Upload file</title>
	<link rel="stylesheet" href="https://gitcdn.link/repo/Chalarangelo/mini.css/master/dist/mini-dark.min.css">
	<style>
		.title {
			border-bottom: 1px dashed rgba(208, 208, 208, 0.1);
		}
		.title span {
			display: inline-block;
			font-size: inherit;
			width: 7rem;
		}
		.title small {
			display: inline-block;
			margin-left: 1rem;
		}
		.title small a {
			margin-right: 0.5rem;
		}
		.title small a:last-child {
			margin-right: 0;
		}
		.input-group [type="checkbox"]+label {
			margin-left: 1.5rem;
		}
	</style>
</head>
<body>
	<div class="container" style="margin-bottom: 2.5rem;">
		<div class="row">
			<div class="col-sm title">
				<h1>
					<span>Upload</span>
					<small>
						<a href="/">Store</a>
						<a href="/stat" target="_blank">Stat</a>
					</small>
				</h1>
			</div>
		</div>
		<div class="row">
			<form action="/url" method="post" class="col-sm">
				<fieldset>
					<legend>File Upload</legend>
					<div class="input-group vertical">
						<label for="uploadfile">uploadfile</label>
						<input type="text" name="uploadfile" id="uploadfile" />
					</div>
					<div class="input-group vertical">
						<label for="pngqlt">pngqlt</label>
						<input type="text" name="pngqlt" id="pngqlt" value="60" />
					</div>
					<div class="input-group vertical">
						<label for="jpgqlt">jpgqlt</label>
						<input type="text" name="jpgqlt" id="jpgqlt" value="75" />
					</div>
					<div class="input-group vertical">
						<input type="checkbox" name="keep" id="keep" value="1" checked />
						<label for="keep">keep</label>
					</div>
				</fieldset>
				<input type="submit" value="Upload" />
			</form>
		</div>
	</div>
</body>
</html>
`

const tpl2 = `
<html>
<head>
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>Upload file</title>
	<link rel="stylesheet" href="https://gitcdn.link/repo/Chalarangelo/mini.css/master/dist/mini-dark.min.css">
	<style>
		.title {
			border-bottom: 1px dashed rgba(208, 208, 208, 0.1);
		}
		.title span {
			display: inline-block;
			font-size: inherit;
			width: 7rem;
		}
		.title small {
			display: inline-block;
			margin-left: 1rem;
		}
		.title small a {
			margin-right: 0.5rem;
		}
		.title small a:last-child {
			margin-right: 0;
		}
		.entry {
			border-bottom: 1px dashed rgba(208, 208, 208, 0.1);
		}
		.entry p {
			word-wrap: break-word;
		}
		.entry p.once {
			position: relative;
		}
		.entry p.once::after {
			background-color: rgba(198, 40, 40, 0.9);
			border-radius: 0.25rem;
			color: #242f33;
			content: "once";
			display: inline-block;
			font-size: 0.7rem;
			font-weight: bold;
			line-height: 1;
			padding: 0.15rem 0 0.25rem;
			position: relative;
			text-align: center;
			top: -0.55rem;
			width: 2.25rem;
		}
		.entry a {
			display: inline-block;
			text-decoration: none;
			text-overflow: ellipsis;
			overflow: hidden;
			white-space: nowrap;
			width: 100%;
		}
		.entry p.once a {
			width: calc(100% - 2.5rem);
		}
	</style>
</head>
<body>
	<div class="container" style="margin-bottom: 2.5rem;">
		<div class="row">
			<div class="col-sm title">
				<h1>
					<span>Store</span>
					<small>
						<a href="/url">Upload</a>
						<a href="/stat" target="_blank">Stat</a>
					</small>
				</h1>
			</div>
		</div>
		<div class="row">
		{{range .}}
			<div class="col-sm-6 col-md-3 col-lg-2 entry">
				<p title="{{.Name}}" {{if eq .Keep 0}}class="once"{{end}}>
					<small>{{.Host}}:</small>
					<br />
					<a href="/store/{{.ID.Hex}}">{{.Name}}</a>
				</p>
			</div>
		{{end}}
		</div>
	</div>
	<script type="text/javascript">
		document
			.querySelectorAll(".once")
			.forEach(function (elm) {
				elm
					.children[2]
					.addEventListener(
						"click",
						function (evt) {
							evt
								.target
								.parentElement
								.parentElement
								.remove();
						}
					);
			});
	</script>
</body>
</html>
`

// Entry is a file entry type
type Entry struct {
	ID          bson.ObjectId `bson:"_id,omitempty"`
	Name        string        `bson:"name"`
	Path        string        `bson:"path"`
	ContentType string        `bson:"content_type"`
	InSize      int64         `bson:"in_size"`
	OutSize     int64         `bson:"out_size"`
	UserName    string        `bson:"user_name"`
	Host        string        `bson:"host"`
	Keep        int8          `bson:"keep"`
	Timestamp   time.Time     `bson:"timestamp"`
}

// Statistic is a statistic type
type Statistic struct {
	ID             bson.ObjectId `bson:"_id,omitempty"`
	UserName       string        `bson:"user_name"`
	Host           string        `bson:"host"`
	UploadQuantity int64         `bson:"upload_quantity"`
	InSize         int64         `bson:"in_size"`
	OutSize        int64         `bson:"out_size"`
}

// User is a user type
type User struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	Name     string        `bson:"name"`
	Password string        `bson:"password"`
}

type contextKey string

func (c contextKey) String() string {
	return string(c)
}

func createAuthMiddleware(realm string, s *mgo.Session) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			name, password, ok := r.BasicAuth()
			if !ok {
				unauthorized(w, realm)
				return
			}

			session := s.Copy()
			defer session.Close()

			u := User{}
			err := session.DB("store").C("users").Find(
				bson.M{
					"name":     name,
					"password": password,
				},
			).One(&u)
			if err != nil {
				unauthorized(w, realm)
				return
			}

			un := "UserID"
			uid := u.ID.Hex()
			if rc, _ := r.Cookie(un); rc == nil || rc.Value != uid {
				c := http.Cookie{
					Name:    un,
					Value:   uid,
					Expires: time.Now().Add(24 * time.Hour),
				}
				http.SetCookie(w, &c)
			}

			key := "User"
			ctx := context.WithValue(r.Context(), contextKey(key), u)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func unauthorized(w http.ResponseWriter, realm string) {
	w.Header().Add("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
	w.WriteHeader(http.StatusUnauthorized)
}

func calcPath(fn string) string {
	crutime := time.Now().Unix()
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%s + %d + %s", strconv.FormatInt(crutime, 20), rand.Int63(), fn))
	hash := h.Sum(nil)

	return fmt.Sprintf("./store/%x-%s", hash, fn)
}

func optImg(pt string, ct string, pngqlt int, jpgqlt int) error {
	if ct == "image/png" {
		cmd := exec.Command("pngquant",
			"--quality",
			fmt.Sprintf("%s-%s", strconv.Itoa(pngqlt), strconv.Itoa(pngqlt)),
			"--ext",
			".png",
			"--force",
			pt,
		)
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	if ct == "image/jpeg" {
		cmd := exec.Command(
			"jpegoptim",
			"-m",
			strconv.Itoa(jpgqlt),
			pt,
		)
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func urlGet(w http.ResponseWriter, r *http.Request) {
	t, _ := template.New("upload").Parse(tpl1)
	t.Execute(w, nil)
}

func urlPost(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		u := r.Context().Value(contextKey("User")).(User)

		jpgqlt, err := strconv.Atoi(r.Form.Get("jpgqlt"))
		if err != nil || jpgqlt < 0 || jpgqlt > 100 {
			jpgqlt = 75
		}

		pngqlt, err := strconv.Atoi(r.Form.Get("pngqlt"))
		if err != nil || pngqlt < 0 || pngqlt > 100 {
			pngqlt = 60
		}

		keep, err := strconv.ParseInt(r.Form.Get("keep"), 10, 64)
		if err != nil {
			keep = 0
		}

		uri := r.Form.Get("uploadfile")
		if _, err = url.ParseRequestURI(uri); err != nil {
			errStr := fmt.Sprintf("ParseRequestURI: %s", err.Error())
			log.Println(errStr)
			http.Error(w, errStr, 400)
			return
		}

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}

		resp, err := client.Get(uri)
		if err != nil {
			errStr := fmt.Sprintf("ClientGet: %s", err.Error())
			log.Println(errStr)
			http.Error(w, errStr, 400)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode > 399 {
			errStr := fmt.Sprintf("ResponseStatusCode: %d", resp.StatusCode)
			log.Println(errStr)
			http.Error(w, errStr, 400)
			return
		}

		if resp.ContentLength > (32 << 19) {
			errStr := fmt.Sprintf("ResponseContentLength: %d", resp.ContentLength)
			log.Println(errStr)
			http.Error(w, errStr, 400)
			return
		}

		ct := resp.Header.Get("Content-Type")
		host := strings.Split(resp.Request.URL.Host, ":")[0]

		sl := strings.Split(resp.Request.URL.Path, "/")
		sl = strings.Split(sl[len(sl)-1], ".")

		fn := strings.Join(sl[0:len(sl)-1], ".")
		if len(fn) == 0 {
			fn = "untitled"
		}

		var ex string
		if len(sl) > 1 {
			ex = fmt.Sprintf(".%s", sl[len(sl)-1])
		}

		if len(ex) == 0 {
			exs, _ := mime.ExtensionsByType(ct)
			if len(exs) > 0 {
				ex = exs[0]
			}
		}

		fn = strings.Join([]string{fn, ex}, "")
		pt := calcPath(fn)

		fo, err := os.OpenFile(pt, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			errStr := fmt.Sprintf("OpenFile: %s", err.Error())
			log.Println(errStr)
			http.Error(w, errStr, 500)
			return
		}
		defer fo.Close()

		io.Copy(fo, resp.Body)
		fii, _ := os.Stat(fo.Name())

		err = optImg(fo.Name(), ct, pngqlt, jpgqlt)
		if err != nil {
			errStr := fmt.Sprintf("OptImg %s: %s", ct, err.Error())
			log.Println(errStr)
			http.Error(w, errStr, 400)

			os.Remove(fo.Name())
			return
		}

		foi, _ := os.Stat(fo.Name())

		sc := s.Copy()
		defer sc.Close()

		st := Statistic{}
		if u.Name == "test" && u.Password == "test" {
			err = sc.DB("store").C("statistics").Find(
				bson.M{
					"host":      host,
					"user_name": u.Name,
				},
			).One(&st)
			tq, _ := strconv.ParseInt(os.Getenv("TEST_QUANTITY"), 10, 64)
			if err == nil && st.UploadQuantity >= tq {
				errStr := fmt.Sprintf("TestQuantityExceeded: %d", st.UploadQuantity)
				log.Println(errStr)
				http.Error(w, errStr, 402)
				return
			}
		}

		id := bson.NewObjectId()
		e := &Entry{
			ID:          id,
			Name:        fn,
			Path:        fo.Name(),
			ContentType: ct,
			InSize:      fii.Size(),
			OutSize:     foi.Size(),
			UserName:    u.Name,
			Host:        host,
			Keep:        int8(keep),
			Timestamp:   time.Now(),
		}
		err = sc.DB("store").C("entries").Insert(e)
		if err != nil {
			errStr := fmt.Sprintf("DbEntriesInsert: %s", err.Error())
			log.Println(errStr)
			http.Error(w, errStr, 500)
			return
		}

		_, err = sc.DB("store").C("statistics").Upsert(
			bson.M{"host": host, "user_name": u.Name},
			bson.M{
				"$set": bson.M{
					"host":      e.Host,
					"user_name": u.Name,
				},
				"$inc": bson.M{
					"upload_quantity": 1,
					"in_size":         e.InSize,
					"out_size":        e.OutSize,
				},
			},
		)
		if err != nil {
			errStr := fmt.Sprintf("DbStatisticsUpsert: %s", err.Error())
			log.Println(errStr)
			http.Error(w, errStr, 500)
			return
		}

		respBody, err := json.MarshalIndent(e, "", "  ")
		if err != nil {
			errStr := fmt.Sprintf("MarshalIndent: %s", err.Error())
			log.Println(errStr)
			http.Error(w, errStr, 500)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(respBody)
	}
}

func serveFiles(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		id := chi.URLParam(r, "id")
		if ch := bson.IsObjectIdHex(id); !ch {
			errStr := fmt.Sprintf("IsObjectIdHex: %t", ch)
			log.Println(errStr)
			http.Error(w, errStr, 400)
			return
		}

		e := Entry{}
		err := session.DB("store").C("entries").Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&e)
		if err != nil {
			errStr := fmt.Sprintf("DbEntriesFindOne: %s", err.Error())
			log.Println(errStr)
			http.Error(w, errStr, 404)
			return
		}

		f, err := os.OpenFile(e.Path, os.O_RDONLY, 0666)
		if err != nil {
			errStr := fmt.Sprintf("OpenFile: %s", err.Error())
			log.Println(errStr)
			http.Error(w, errStr, 404)

			err := session.DB("store").C("entries").RemoveId(e.ID)
			if err != nil {
				errStr := fmt.Sprintf("DbEntriesRemoveId: %s", err.Error())
				log.Println(errStr)
				http.Error(w, errStr, 500)
				return
			}
			return
		}
		defer f.Close()

		w.Header().Set("Content-Disposition", "attachment; filename="+e.Name)
		w.Header().Set("Content-Type", e.ContentType)

		io.Copy(w, f)

		if e.Keep == 0 {
			os.Remove(e.Path)
			err := session.DB("store").C("entries").RemoveId(e.ID)
			if err != nil {
				errStr := fmt.Sprintf("DbEntriesRemoveId: %s", err.Error())
				log.Println(errStr)
				http.Error(w, errStr, 500)
				return
			}
		}
	}
}

func showFiles(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		u := r.Context().Value(contextKey("User")).(User)

		var es []Entry
		err := session.DB("store").C("entries").Find(bson.M{"user_name": u.Name}).Sort("-timestamp").Limit(500).All(&es)
		if err != nil {
			errStr := fmt.Sprintf("DbEntriesFindAll: %s", err.Error())
			log.Println(errStr)
			http.Error(w, errStr, 500)
			return
		}

		t, _ := template.New("upload").Parse(tpl2)
		err = t.Execute(w, es)
		if err != nil {
			errStr := fmt.Sprintf("TemplateExecute: %s", err.Error())
			log.Println(errStr)
			http.Error(w, errStr, 500)
			return
		}
	}
}

func statisticGet(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		u := r.Context().Value(contextKey("User")).(User)

		session := s.Copy()
		defer session.Close()

		var sts []Statistic
		err := session.DB("store").C("statistics").Find(bson.M{"user_name": u.Name}).All(&sts)
		if err != nil {
			errStr := fmt.Sprintf("DbStatisticsFindAll: %s", err.Error())
			log.Println(errStr)
			http.Error(w, errStr, 500)
			return
		}

		respBody, err := json.MarshalIndent(sts, "", "  ")
		if err != nil {
			errStr := fmt.Sprintf("MarshalIndent: %s", err.Error())
			log.Println(errStr)
			http.Error(w, errStr, 500)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(respBody)
	}
}

func statisticPost(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		u := r.Context().Value(contextKey("User")).(User)

		host := r.Form.Get("host")

		session := s.Copy()
		defer session.Close()

		st := Statistic{}
		err := session.DB("store").C("statistics").Find(
			bson.M{
				"user_name": u.Name,
				"host":      host,
			},
		).One(&st)
		if err != nil {
			errStr := fmt.Sprintf("DbStatisticsFindOne: %s", err.Error())
			log.Println(errStr)
			http.Error(w, errStr, 400)
			return
		}

		respBody, err := json.MarshalIndent(st, "", "  ")
		if err != nil {
			errStr := fmt.Sprintf("MarshalIndent: %s", err.Error())
			log.Println(errStr)
			http.Error(w, errStr, 500)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(respBody)
	}
}

func cleanGet(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		var dtk int
		dtk, _ = strconv.Atoi(os.Getenv("DAYS_TO_KEEP"))

		var es []Entry
		err := session.DB("store").C("entries").Find(
			bson.M{
				"$or": []bson.M{
					bson.M{
						"$and": []bson.M{
							bson.M{
								"user_name": "test",
							},
							bson.M{
								"timestamp": bson.M{
									"$lt": time.Now().Add(-time.Minute),
								},
							},
						},
					},
					bson.M{
						"timestamp": bson.M{
							"$lt": time.Now().AddDate(0, 0, -dtk),
						},
					},
				},
			},
		).All(&es)
		if err != nil {
			errStr := fmt.Sprintf("DbEntriesFindAll: %s", err.Error())
			log.Println(errStr)
			http.Error(w, errStr, 400)
			return
		}

		for _, e := range es {
			os.Remove(e.Path)
			err := session.DB("store").C("entries").RemoveId(e.ID)
			if err != nil {
				errStr := fmt.Sprintf("DbEntriesRemoveId: %s", err.Error())
				log.Println(errStr)
				http.Error(w, errStr, 500)
				return
			}
		}

		resStr := fmt.Sprintf("Cleaned: %d", len(es))

		io.WriteString(w, resStr)
		log.Println(resStr)
	}
}

func main() {
	mgoAddr := os.Getenv("DB_ADDR")
	session, err := mgo.Dial(mgoAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	r := chi.NewRouter()

	authUserMiddleware := createAuthMiddleware("MyRealm", session)

	r.Route("/", func(r chi.Router) {
		r.With(authUserMiddleware).Get("/", showFiles(session))

		r.Get("/store/{id}", serveFiles(session))

		r.Get("/clean", cleanGet(session))

		r.Route("/url", func(r chi.Router) {
			r.Use(authUserMiddleware)
			r.Get("/", urlGet)
			r.Post("/", urlPost(session))
		})

		r.Route("/stat", func(r chi.Router) {
			r.Use(authUserMiddleware)
			r.Get("/", statisticGet(session))
			r.Post("/", statisticPost(session))
		})
	})

	servAddr := strings.Join([]string{os.Getenv("HOST"), os.Getenv("PORT")}, ":")
	err = http.ListenAndServe(servAddr, r)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
