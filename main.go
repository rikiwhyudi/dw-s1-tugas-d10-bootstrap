package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
	"tugas10/connection"

	"github.com/gorilla/mux"
)

// var Data = map[string]interface{} {
// 	"Title": "Personal Web",
// }

//tipe data penampung hasil query
type Blog struct {
	Id				int
	Title 			string
	Images			string
	Start_date 		time.Time
	End_date		time.Time
	Duration		string
	// Post_date 	time.Time
	SFormat_date	string
	EFormat_date	string
	Author			string
	// Technologies 	string
	Content 		string
}

var Blogs = []Blog{
	
}

func main() {

	route := mux.NewRouter()

	connection.DatabaseConnect()

	// route path folder untuk public
	route.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer((http.Dir("./public")))))

	//routing. parameter pertama adalah rute dan parameter ke-2 adalah handlernya dengan method get dan post dll
	route.HandleFunc("/", home).Methods("GET")
	route.HandleFunc("/contact", contact).Methods("GET")
	route.HandleFunc("/blog-detail/{id}", blogDetail).Methods("GET")
	route.HandleFunc("/blog", form).Methods("GET")
	route.HandleFunc("/process", process).Methods("POST")
	route.HandleFunc("/delete/{id}", deleted).Methods("GET")


	fmt.Println("Server running on port 5000");
	//membuat sekaligus start server baru
	http.ListenAndServe("localhost:5000", route)
}

	/* untuk keperluan penanganan request ke rute yang ditentukan
	 Parameter ke-1 merupakan objek untuk keperluan http response
	 parameter ke-2 yang bertipe *request, berisikan informasi-informasi yang berhubungan dengan http request untuk rute yang bersangkutan.
	 */
func home(w http.ResponseWriter, r *http.Request) {
	//mengatur header
	w.Header().Set("Content-Type", "text/html; charset=utf-8")


	//membuat variabel memparsing template halaman index
	var tmpl, err  = template.ParseFiles("views/index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		//meng-output-kan nilai balik data. Argumen method adalah data yang ingin dijadikan output
		w.Write([]byte("message : " + err.Error()))
		return
	}

	//mengambil semua data yg di select dari database di tb_project untuk kemudian di render ke halaman depan (index).
	rows, _ := connection.Conn.Query(context.Background(), "SELECT id, name, start_date, end_date, description, image, author FROM tb_projects")

	var result []Blog //data slice of array di gunakan untuk menampung hasil query

	for rows.Next() {
		var each = Blog{} //memanggil struct
		//scan mengambil nilai record yang sedang diiterasi, untuk disimpan pada variabel pointer
		err := rows.Scan(&each.Id, &each.Title, &each.Start_date, &each.End_date, &each.Content, &each.Images, &each.Author)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		each.Author = "Riki Wahyudi" //optional if you dont have author in db, you can comment this if you want
		each.Duration = getDuration(each.Start_date, each.End_date)
		each.SFormat_date = each.Start_date.Format("2 January 2006") //Post_at temporary
		result = append(result, each)
	}
	fmt.Println(result)

	respData := map[string]interface{}{
		"Blogs": result,
	}	

	w.WriteHeader(http.StatusOK) 
	tmpl.Execute(w, respData)
}

func blogDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	//konversi string ke int
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	var tmpl, err = template.ParseFiles("views/blog-detail.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message :" +err.Error()))
		return
	}

	var BlogDetail = Blog{}
	/*mengambil data dari letak id yg di pilih dari database di tb_project untuk kemudian di render ke halaman details blog.
	  berdasarkan id nya
	*/
	err = connection.Conn.QueryRow(context.Background(), "SELECT id, name, start_date, end_date, description, image, author FROM tb_projects WHERE id=$1", id).Scan(&BlogDetail.Id, &BlogDetail.Title, &BlogDetail.Start_date, &BlogDetail.End_date, &BlogDetail.Content, &BlogDetail.Images, &BlogDetail.Author)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	BlogDetail.Duration = getDuration(BlogDetail.Start_date, BlogDetail.End_date)
	BlogDetail.SFormat_date = BlogDetail.Start_date.Format("2 January 2006")
	BlogDetail.EFormat_date = BlogDetail.End_date.Format("2 January 2006")

	data := map[string]interface{}{
		"Blog": BlogDetail,
	}
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, data)
}

func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var  tmpl, err = template.ParseFiles("views/form.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message :" +err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, nil)
}

func form(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var tmpl, err = template.ParseFiles("views/blog.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message :" +err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, nil)

}

func process(w http.ResponseWriter, r *http.Request) {
	// w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println("Title :" + r.PostForm.Get("inputTitle"))
	// fmt.Println("Content :" + r.PostForm.Get("inputContent"))
	// fmt.Println("StartDate :" + r.PostForm.Get("inputStart"))
	// fmt.Println("EndDate :" + r.PostForm.Get("inputEnd"))

	var title = r.PostForm.Get("inputTitle")
	var content = r.PostForm.Get("inputContent")
	var start = r.PostForm.Get("inputStart")
	var end = r.PostForm.Get("inputEnd")

	//static image and author after create post to insert in db
	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_projects(name, start_date, end_date, description, image, author) VALUES ($1, $2, $3, $4, 'work-unsplash.jpg', 'Riki Wahyudi')", title, start, end, content)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " +err.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func deleted(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	_, err := connection.Conn.Exec(context.Background(), "DELETE FROM tb_projects WHERE id=$1", id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func getDuration(Start_date time.Time, End_date time.Time) string {

	timeFormat := "2006-01-02"

	start, _ := time.Parse(timeFormat, Start_date.Format(timeFormat))
	end, _ := time.Parse(timeFormat, End_date.Format(timeFormat))

	distance := end.Sub(start).Hours() / 24
	var duration string

	if distance > 30 {
		if (distance / 30) <= 1 {
			duration = "1 Month"
		}
	duration = strconv.Itoa(int(distance)/30) + " Month"
	} else {
		if distance <= 1 {
			duration = "1 Days"
		} 
	duration = strconv.Itoa(int(distance)) + " Days"
	}

	return duration
}