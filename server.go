package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	_ "github.com/mattn/go-sqlite3"
)

type user struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Hod  int    `json:"hod"`
}

type req_user struct {
	Hod int `json:"hod"`
}

type meeting struct {
	Id      int    `json:"id"`
	Label   string `json:"label"`
	Creator user   `json:"creator"`
	Slot    string `json:"slot"`
	Users   []user `json:"users"`
}

type req_meeting struct {
	Label string   `json:"label"`
	Slot  string   `json:"slot"`
	Users []string `json:"users"`
}

type blocked struct {
	Id    int    `json:"id"`
	Label string `json:"label"`
	User  user   `json:"user"`
	Slot  string `json:"slot"`
}

type req_blocked struct {
	Label string `json:"label"`
	Slot  string `json:"slot"`
}

type all_slots struct {
	User     user      `json:"user"`
	Meetings []meeting `json:"meetings"`
	Blocks   []blocked `json:"blocks"`
}

type all_slots_hod struct {
	User             user      `json:"user"`
	Meetings         []meeting `json:"meetings"`
	Blocks           []blocked `json:"blocks"`
	Monthly_Meetings []meeting `json:"monthly_meetings"`
}

func main() {

	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createStmt := `
    create table if not exists user (id integer not null primary key autoincrement, name text not null unique, hod integer default 0);
    create table if not exists meeting (id integer not null primary key autoincrement, creator_id integer not null, user_id integer not null, slot text not null, label text not null);
    create table if not exists blocked (id integer not null primary key autoincrement, user_id integer not null, slot text not null, label text not null);
    `
	_, err = db.Exec(createStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, createStmt)
		return
	}

	var validPath = regexp.MustCompile("^/([a-zA-Z0-9]+)/*(|blockCalendar|scheduleMeeting)$")
	// var validName = regexp.MustCompile("^([a-zA-Z0-9]+)$")
	var validDate = regexp.MustCompile("^(\\d{4})-(\\d{2})-(\\d{2})(T| )([0-9]|0[0-9]|1[0-9]|2[0-3]):00(|:00)(Z|)$")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		fmt.Println(len(m))
		if m == nil {
			fmt.Fprintf(w, "{\"message\":\"Your path is invalid. Please check the documentation for usage.\",\"success\":false}")
			return
		}
		if r.Method == "GET" {
			if m[2] == "" {
				var user1 user
				rows, err := db.Query(`SELECT * FROM user WHERE name = '` + m[1] + `';`)
				if err != nil {
					fmt.Fprintf(w, "%q\n", err)
					return
				} else if rows.Next() {
					rows.Scan(&user1.Id, &user1.Name, &user1.Hod)
					if user1.Hod == 0 {
						var response all_slots
						response.User = user1
						rows2, err2 := db.Query(`SELECT * FROM blocked WHERE user_id = ` + fmt.Sprint(user1.Id) + ` AND slot BETWEEN date('now', 'start of day') AND date('now', 'start of day', '+1 day');`)
						if err2 != nil {
							fmt.Fprintf(w, "%q\n", err2)
							return
						}
						var blocks []blocked
						var uid int
						for rows2.Next() {
							var block blocked
							rows2.Scan(&block.Id, &uid, &block.Slot, &block.Label)
							block.User = user1
							blocks = append(blocks, block)
						}
						response.Blocks = blocks
						rows3, err3 := db.Query(`SELECT * FROM meeting WHERE user_id = ` + fmt.Sprint(user1.Id) + ` AND slot BETWEEN date('now', 'start of day') AND date('now', 'start of day', '+1 day');`)
						if err3 != nil {
							fmt.Fprintf(w, "%q\n", err3)
							return
						}
						var meets []meeting
						for rows3.Next() {
							var meet meeting
							var cid int
							rows3.Scan(&meet.Id, &cid, &uid, &meet.Slot, &meet.Label)
							rows4, err4 := db.Query(`SELECT * FROM user WHERE id = ` + fmt.Sprint(cid) + `;`)
							if err4 != nil {
								fmt.Fprintf(w, "%q\n", err4)
								return
							} else if rows4.Next() {
								var creator user
								rows4.Scan(&creator.Id, &creator.Name, &creator.Hod)
								meet.Creator = creator
							}
							rows5, err5 := db.Query(`SELECT user.id,user.name,user.hod FROM user JOIN meeting ON user.id=meeting.user_id WHERE meeting.creator_id = ` + fmt.Sprint(cid) + ` AND datetime(meeting.slot) = datetime('` + meet.Slot + `')`)
							if err5 != nil {
								fmt.Fprintf(w, "%q\n", err5)
								return
							}
							var users []user
							for rows5.Next() {
								var userm user
								rows5.Scan(&userm.Id, &userm.Name, &userm.Hod)
								users = append(users, userm)
							}
							meet.Users = users
							meets = append(meets, meet)
						}
						response.Blocks = blocks
						response.Meetings = meets
						json.NewEncoder(w).Encode(response)
					} else {
						var response all_slots_hod
						response.User = user1
						rows2, err2 := db.Query(`SELECT * FROM blocked WHERE user_id = ` + fmt.Sprint(user1.Id) + ` AND slot BETWEEN date('now', 'start of day') AND date('now', 'start of day', '+1 day');`)
						if err2 != nil {
							fmt.Fprintf(w, "%q\n", err2)
							return
						}
						var blocks []blocked
						var uid int
						for rows2.Next() {
							var block blocked
							rows2.Scan(&block.Id, &uid, &block.Slot, &block.Label)
							block.User = user1
							blocks = append(blocks, block)
						}
						response.Blocks = blocks
						rows3, err3 := db.Query(`SELECT * FROM meeting WHERE user_id = ` + fmt.Sprint(user1.Id) + ` AND slot BETWEEN date('now', 'start of day') AND date('now', 'start of day', '+1 day');`)
						if err3 != nil {
							fmt.Fprintf(w, "%q\n", err3)
							return
						}
						var meets []meeting
						for rows3.Next() {
							var meet meeting
							var cid int
							rows3.Scan(&meet.Id, &cid, &uid, &meet.Slot, &meet.Label)
							rows4, err4 := db.Query(`SELECT * FROM user WHERE id = ` + fmt.Sprint(cid) + `;`)
							if err4 != nil {
								fmt.Fprintf(w, "%q\n", err4)
								return
							} else if rows4.Next() {
								var creator user
								rows4.Scan(&creator.Id, &creator.Name, &creator.Hod)
								meet.Creator = creator
							}
							rows5, err5 := db.Query(`SELECT user.id,user.name,user.hod FROM user JOIN meeting ON user.id=meeting.user_id WHERE meeting.creator_id = ` + fmt.Sprint(cid) + ` AND datetime(meeting.slot) = datetime('` + meet.Slot + `')`)
							if err5 != nil {
								fmt.Fprintf(w, "%q\n", err5)
								return
							}
							var users []user
							for rows5.Next() {
								var userm user
								rows5.Scan(&userm.Id, &userm.Name, &userm.Hod)
								users = append(users, userm)
							}
							meet.Users = users
							meets = append(meets, meet)
						}
						rows6, err6 := db.Query(`SELECT * FROM meeting WHERE slot BETWEEN date('now', 'start of month') AND date('now', 'start of month','+1 month');`)
						if err6 != nil {
							fmt.Fprintf(w, "%q\n", err6)
							return
						}
						var m_meets []meeting
						for rows6.Next() {
							var meet meeting
							var cid int
							rows6.Scan(&meet.Id, &cid, &uid, &meet.Slot, &meet.Label)
							rows4, err4 := db.Query(`SELECT * FROM user WHERE id = ` + fmt.Sprint(cid) + `;`)
							if err4 != nil {
								fmt.Fprintf(w, "%q\n", err4)
								return
							} else if rows4.Next() {
								var creator user
								rows4.Scan(&creator.Id, &creator.Name, &creator.Hod)
								meet.Creator = creator
							}
							rows5, err5 := db.Query(`SELECT user.id,user.name,user.hod FROM user JOIN meeting ON user.id=meeting.user_id WHERE meeting.creator_id = ` + fmt.Sprint(cid) + ` AND datetime(meeting.slot) = datetime('` + meet.Slot + `')`)
							if err5 != nil {
								fmt.Fprintf(w, "%q\n", err5)
								return
							}
							var users []user
							for rows5.Next() {
								var userm user
								rows5.Scan(&userm.Id, &userm.Name, &userm.Hod)
								users = append(users, userm)
							}
							meet.Users = users
							m_meets = append(m_meets, meet)
						}
						response.Blocks = blocks
						response.Meetings = meets
						response.Monthly_Meetings = m_meets
						json.NewEncoder(w).Encode(response)
					}
				} else {
					fmt.Fprintf(w, "{\"message\":\"User not found.\",\"success\":false}")
					return
				}
			} else if m[2] == "blockCalendar" {
				fmt.Fprintf(w, "blockCalendar Operation endpoint")

			} else if m[2] == "scheduleMeeting" {
				fmt.Fprintf(w, "scheduleMeeting Operation endpoint")
			} else {
				fmt.Fprintf(w, "Unknown Operation")
			}
		} else if r.Method == "POST" {
			if m[2] == "" {
				var user1 req_user
				reqBody, _ := ioutil.ReadAll(r.Body)

				err0 := json.Unmarshal(reqBody, &user1)
				if err0 != nil {
					fmt.Fprintf(w, "{\"message\":\"Please supply a json body. "+err0.Error()+"\",\"success\":false}")
					return
				}
				rows, err := db.Query(`SELECT * FROM user WHERE name = '` + m[1] + `';`)
				defer rows.Close()
				if err != nil {
					fmt.Fprintf(w, "%q\n", err)
					return
				} else if rows.Next() {
					fmt.Fprintf(w, "{\"message\":\"User already exists.\",\"success\":false}")
					return
				} else {
					if user1.Hod == 0 || user1.Hod == 1 {
						_, err2 := db.Exec(`INSERT INTO user (name,hod) VALUES("` + m[1] + `","` + fmt.Sprint(user1.Hod) + `")`)
						if err2 != nil {
							log.Printf("%q\n", err2)
							return
						} else {
							fmt.Fprintf(w, "{\"message\":\"User '"+m[1]+"' created successfully.\",\"success\":true}")
							return
						}
					} else {
						fmt.Fprintf(w, "{\"message\":\"Invalid value for the field hod. hod=(0|1)\",\"success\":false}")
						return
					}
				}

			} else if m[2] == "blockCalendar" {
				var block req_blocked
				reqBody, _ := ioutil.ReadAll(r.Body)

				err0 := json.Unmarshal(reqBody, &block)
				if err0 != nil {
					fmt.Fprintf(w, "{\"message\":\"Please supply a json body. "+err0.Error()+"\",\"success\":false}")
					return
				}
				rows, err := db.Query(`SELECT * FROM user WHERE name = '` + m[1] + `';`)
				if err != nil {
					fmt.Fprintf(w, "%q\n", err)
					return
				} else if rows.Next() {
					var user1 user
					rows.Scan(&user1.Id, &user1.Name, &user1.Hod)
					rows.Close()
					if validDate.MatchString(block.Slot) {
						rows2, err2 := db.Query(`SELECT * FROM blocked WHERE user_id=` + fmt.Sprint(user1.Id) + ` AND datetime(slot) = datetime('` + block.Slot + `')`)
						rows3, err3 := db.Query(`SELECT * FROM meeting WHERE user_id=` + fmt.Sprint(user1.Id) + ` AND datetime(slot) = datetime('` + block.Slot + `')`)
						defer rows2.Close()
						defer rows3.Close()
						if err2 != nil || err3 != nil {
							log.Printf("%q\n", err2)
							return
						}
						if rows2.Next() {
							fmt.Fprintf(w, "{\"message\":\"Slot is already reserved (blocked) for '"+m[1]+"'.\",\"success\":false}")
							return
						} else if rows3.Next() {
							fmt.Fprintf(w, "{\"message\":\"Slot is already reserved (meeting) for '"+m[1]+"'.\",\"success\":false}")
							return
						} else {
							_, err4 := db.Exec(`INSERT INTO blocked (user_id,slot,label) VALUES(` + fmt.Sprint(user1.Id) + `,'` + block.Slot + `','` + block.Label + `')`)
							if err4 != nil {
								log.Printf("%q\n", err4)
								return
							} else {
								fmt.Fprintf(w, "{\"message\":\"Slot ("+block.Slot+") labelled '"+block.Label+"' blocked successfully for '"+m[1]+"'.\",\"success\":true}")
								return
							}
						}
					} else {
						fmt.Fprintf(w, "{\"message\":\"Invalid value for slot.\",\"success\":false}")
						return
					}
				} else {
					fmt.Fprintf(w, "{\"message\":\"User doesnot exists.\",\"success\":false}")
					return
				}

			} else if m[2] == "scheduleMeeting" {
				var meet req_meeting
				reqBody, _ := ioutil.ReadAll(r.Body)

				err0 := json.Unmarshal(reqBody, &meet)
				if err0 != nil {
					fmt.Fprintf(w, "{\"message\":\"Please supply a json body. "+err0.Error()+"\",\"success\":false}")
					return
				}
				rows, err := db.Query(`SELECT * FROM user WHERE name = '` + m[1] + `';`)
				if err != nil {
					fmt.Fprintf(w, "%q\n", err)
					return
				} else if rows.Next() {
					var user1 user
					rows.Scan(&user1.Id, &user1.Name, &user1.Hod)
					rows.Close()
					if validDate.MatchString(meet.Slot) {
						rows2, err2 := db.Query(`SELECT * FROM blocked WHERE user_id=` + fmt.Sprint(user1.Id) + ` AND datetime(slot) = datetime('` + meet.Slot + `')`)
						rows3, err3 := db.Query(`SELECT * FROM meeting WHERE user_id=` + fmt.Sprint(user1.Id) + ` AND datetime(slot) = datetime('` + meet.Slot + `')`)
						defer rows2.Close()
						defer rows3.Close()
						if err2 != nil || err3 != nil {
							log.Printf("%q\n%q\n", err2, err3)
							return
						}
						if rows2.Next() {
							fmt.Fprintf(w, "{\"message\":\"Slot is already reserved (blocked) for '"+m[1]+"'.\",\"success\":false}")
							return
						} else if rows3.Next() {
							fmt.Fprintf(w, "{\"message\":\"Slot is already reserved (meeting) for '"+m[1]+"'.\",\"success\":false}")
							return
						} else {
							var users []user
							for _, u := range meet.Users {
								rowsu, erru := db.Query(`SELECT * FROM user WHERE name='` + u + `'`)
								if erru != nil {
									log.Printf("%q\n", erru)
									return
								}
								if rowsu.Next() {
									var useru user
									rowsu.Scan(&useru.Id, &useru.Name, &useru.Hod)
									rowsu.Close()
									rowsu2, erru2 := db.Query(`SELECT * FROM meeting WHERE user_id=` + fmt.Sprint(useru.Id) + ` AND datetime(slot) = datetime('` + meet.Slot + `')`)
									if erru2 != nil {
										log.Printf("%q\n", erru2)
										return
									}
									if rowsu2.Next() {
										fmt.Fprintf(w, "{\"message\":\"Slot is already reserved (blocked) for '"+u+"'.\",\"success\":false}")
										return
									}
									rowsu2.Close()
									users = append(users, useru)
								} else {
									fmt.Fprintf(w, "{\"message\":\"User "+u+" doesnot exists.\",\"success\":false}")
									return
								}

							}
							_, err4 := db.Exec(`INSERT INTO meeting (creator_id,user_id,slot,label) VALUES(` + fmt.Sprint(user1.Id) + `,` + fmt.Sprint(user1.Id) + `,'` + meet.Slot + `','` + meet.Label + `')`)
							if err4 != nil {
								log.Printf("%q\n", err4)
								return
							} else {
								for _, useru := range users {
									_, erru := db.Exec(`INSERT INTO meeting (creator_id,user_id,slot,label) VALUES(` + fmt.Sprint(user1.Id) + `,` + fmt.Sprint(useru.Id) + `,'` + meet.Slot + `','` + meet.Label + `')`)
									if erru != nil {
										log.Printf("%q\n", erru)
										return
									}
								}
								fmt.Fprintf(w, "{\"message\":\"Meeting set at slot ("+meet.Slot+") labelled'"+meet.Label+"' successfully by '"+m[1]+"' with "+fmt.Sprint(meet.Users)+".\",\"success\":true}")
								return
							}
						}
					} else {
						fmt.Fprintf(w, "{\"message\":\"Invalid value for slot.\",\"success\":false}")
						return
					}
				} else {
					fmt.Fprintf(w, "{\"message\":\"User doesnot exists.\",\"success\":false}")
					return
				}

			} else {
				fmt.Fprintf(w, "Unknown Operation")
			}
		} else if r.Method == "PUT" {
			if m[2] == "" {
				var user1 req_user
				reqBody, _ := ioutil.ReadAll(r.Body)

				err0 := json.Unmarshal(reqBody, &user1)
				if err0 != nil {
					fmt.Fprintf(w, "{\"message\":\"Please supply a json body. "+err0.Error()+"\",\"success\":false}")
					return
				}
				rows, err := db.Query(`SELECT * FROM user WHERE name = '` + m[1] + `';`)
				if err != nil {
					fmt.Fprintf(w, "%q\n", err)
					return
				} else if rows.Next() {
					rows.Close()
					if user1.Hod == 0 || user1.Hod == 1 {
						_, err2 := db.Exec(`UPDATE user SET hod=` + fmt.Sprint(user1.Hod) + ` WHERE name="` + m[1] + `"`)
						if err2 != nil {
							log.Printf("%q\n", err2)
							return
						} else {
							fmt.Fprintf(w, "{\"message\":\"User '"+m[1]+"' updated successfully.\",\"success\":true}")
							return
						}
					} else {
						fmt.Fprintf(w, "{\"message\":\"Invalid value for the field hod. hod=(0|1)\",\"success\":false}")
						return
					}
				} else {
					fmt.Fprintf(w, "{\"message\":\"User doesnot exist.\",\"success\":false}")
					return
				}

			} else if m[2] == "blockCalendar" {
				var block req_blocked
				reqBody, _ := ioutil.ReadAll(r.Body)

				err0 := json.Unmarshal(reqBody, &block)
				if err0 != nil {
					fmt.Fprintf(w, "{\"message\":\"Please supply a json body. "+err0.Error()+"\",\"success\":false}")
					return
				}
				rows, err := db.Query(`SELECT * FROM user WHERE name = '` + m[1] + `';`)
				if err != nil {
					fmt.Fprintf(w, "%q\n", err)
					return
				} else if rows.Next() {
					var user1 user
					rows.Scan(&user1.Id, &user1.Name, &user1.Hod)
					rows.Close()
					if validDate.MatchString(block.Slot) {
						rows2, err2 := db.Query(`SELECT * FROM blocked WHERE user_id=` + fmt.Sprint(user1.Id) + ` AND datetime(slot) = datetime('` + block.Slot + `')`)
						if err2 != nil {
							log.Printf("%q\n", err2)
							return
						}
						if rows2.Next() {
							rows2.Close()
							_, err4 := db.Exec(`UPDATE blocked SET label = '` + block.Label + `' WHERE user_id=` + fmt.Sprint(user1.Id) + ` AND datetime(slot)=datetime('` + block.Slot + `')`)
							if err4 != nil {
								log.Printf("%q\n", err4)
								return
							} else {
								fmt.Fprintf(w, "{\"message\":\"Blocked slot ("+block.Slot+") labelled '"+block.Label+"' updated successfully for '"+m[1]+"'.\",\"success\":true}")
								return
							}
						} else {
							fmt.Fprintf(w, "{\"message\":\"Slot not found (blocked) for '"+m[1]+"'.\",\"success\":false}")
							return
						}
					} else {
						fmt.Fprintf(w, "{\"message\":\"Invalid value for slot.\",\"success\":false}")
						return
					}
				} else {
					fmt.Fprintf(w, "{\"message\":\"User doesnot exists.\",\"success\":false}")
					return
				}

			} else if m[2] == "scheduleMeeting" {
				var meet req_meeting
				reqBody, _ := ioutil.ReadAll(r.Body)

				err0 := json.Unmarshal(reqBody, &meet)
				if err0 != nil {
					fmt.Fprintf(w, "{\"message\":\"Please supply a json body. "+err0.Error()+"\",\"success\":false}")
					return
				}
				rows, err := db.Query(`SELECT * FROM user WHERE name = '` + m[1] + `';`)
				if err != nil {
					fmt.Fprintf(w, "%q\n", err)
					return
				} else if rows.Next() {
					var user1 user
					rows.Scan(&user1.Id, &user1.Name, &user1.Hod)
					rows.Close()
					if validDate.MatchString(meet.Slot) {
						rows3, err3 := db.Query(`SELECT * FROM meeting WHERE creator_id=` + fmt.Sprint(user1.Id) + ` AND user_id=` + fmt.Sprint(user1.Id) + ` AND datetime(slot) = datetime('` + meet.Slot + `')`)
						if err3 != nil {
							log.Printf("%q\n", err3)
							return
						} else if rows3.Next() {
							rows3.Close()
							_, err2 := db.Exec(`DELETE FROM meeting WHERE creator_id=` + fmt.Sprint(user1.Id) + ` AND datetime(slot) = datetime('` + meet.Slot + `')`)
							if err2 != nil {
								log.Printf("%q\n", err2)
								return
							}
							var users []user
							for _, u := range meet.Users {
								rowsu, erru := db.Query(`SELECT * FROM user WHERE name='` + u + `'`)
								if erru != nil {
									log.Printf("%q\n", erru)
									return
								}
								if rowsu.Next() {
									var useru user
									rowsu.Scan(&useru.Id, &useru.Name, &useru.Hod)
									rowsu.Close()
									rowsu2, erru2 := db.Query(`SELECT * FROM meeting WHERE user_id=` + fmt.Sprint(useru.Id) + ` AND datetime(slot) = datetime('` + meet.Slot + `')`)
									if erru2 != nil {
										log.Printf("%q\n", erru2)
										return
									}
									if rowsu2.Next() {
										fmt.Fprintf(w, "{\"message\":\"Slot is already reserved (blocked) for '"+u+"'.\",\"success\":false}")
										return
									}
									rowsu2.Close()
									users = append(users, useru)
								} else {
									fmt.Fprintf(w, "{\"message\":\"User "+u+" doesnot exists.\",\"success\":false}")
									return
								}

							}
							_, err4 := db.Exec(`INSERT INTO meeting (creator_id,user_id,slot,label) VALUES(` + fmt.Sprint(user1.Id) + `,` + fmt.Sprint(user1.Id) + `,'` + meet.Slot + `','` + meet.Label + `')`)
							if err4 != nil {
								log.Printf("%q\n", err4)
								return
							} else {
								for _, useru := range users {
									_, erru := db.Exec(`INSERT INTO meeting (creator_id,user_id,slot,label) VALUES(` + fmt.Sprint(user1.Id) + `,` + fmt.Sprint(useru.Id) + `,'` + meet.Slot + `','` + meet.Label + `')`)
									if erru != nil {
										log.Printf("%q\n", erru)
										return
									}
								}
								fmt.Fprintf(w, "{\"message\":\"Meeting set at slot ("+meet.Slot+") labelled'"+meet.Label+"' successfully by '"+m[1]+"' with "+fmt.Sprint(meet.Users)+".\",\"success\":true}")
								return
							}
						} else {
							fmt.Fprintf(w, "{\"message\":\"Slot (meeting) for '"+m[1]+"' not found.\",\"success\":false}")
							return
						}
					} else {
						fmt.Fprintf(w, "{\"message\":\"Invalid value for slot.\",\"success\":false}")
						return
					}
				} else {
					fmt.Fprintf(w, "{\"message\":\"User doesnot exists.\",\"success\":false}")
					return
				}

			} else {
				fmt.Fprintf(w, "Unknown Operation")
			}
		} else if r.Method == "PATCH" {
			if m[2] == "" {
				var user1 req_user
				reqBody, _ := ioutil.ReadAll(r.Body)

				err0 := json.Unmarshal(reqBody, &user1)
				if err0 != nil {
					fmt.Fprintf(w, "{\"message\":\"Please supply a json body. "+err0.Error()+"\",\"success\":false}")
					return
				}
				rows, err := db.Query(`SELECT * FROM user WHERE name = '` + m[1] + `';`)
				if err != nil {
					fmt.Fprintf(w, "%q\n", err)
					return
				} else if rows.Next() {
					rows.Close()
					if user1.Hod == 0 || user1.Hod == 1 {
						_, err2 := db.Exec(`UPDATE user SET hod=` + fmt.Sprint(user1.Hod) + ` WHERE name="` + m[1] + `"`)
						if err2 != nil {
							log.Printf("%q\n", err2)
							return
						} else {
							fmt.Fprintf(w, "{\"message\":\"User '"+m[1]+"' updated successfully.\",\"success\":true}")
							return
						}
					} else {
						fmt.Fprintf(w, "{\"message\":\"Invalid value for the field hod. hod=(0|1)\",\"success\":false}")
						return
					}
				} else {
					fmt.Fprintf(w, "{\"message\":\"User doesnot exist.\",\"success\":false}")
					return
				}
			} else if m[2] == "blockCalendar" {
				var block req_blocked
				reqBody, _ := ioutil.ReadAll(r.Body)

				err0 := json.Unmarshal(reqBody, &block)
				if err0 != nil {
					fmt.Fprintf(w, "{\"message\":\"Please supply a json body. "+err0.Error()+"\",\"success\":false}")
					return
				}
				rows, err := db.Query(`SELECT * FROM user WHERE name = '` + m[1] + `';`)
				if err != nil {
					fmt.Fprintf(w, "%q\n", err)
					return
				} else if rows.Next() {
					var user1 user
					rows.Scan(&user1.Id, &user1.Name, &user1.Hod)
					rows.Close()
					if validDate.MatchString(block.Slot) {
						rows2, err2 := db.Query(`SELECT * FROM blocked WHERE user_id=` + fmt.Sprint(user1.Id) + ` AND datetime(slot) = datetime('` + block.Slot + `')`)
						if err2 != nil {
							log.Printf("%q\n", err2)
							return
						}
						if rows2.Next() {
							rows2.Close()
							if block.Label != "" {
								_, err4 := db.Exec(`UPDATE blocked SET label = '` + block.Label + `' WHERE user_id=` + fmt.Sprint(user1.Id) + ` AND datetime(slot)=datetime('` + block.Slot + `')`)
								if err4 != nil {
									log.Printf("%q\n", err4)
									return
								}
							}
							fmt.Fprintf(w, "{\"message\":\"Blocked slot ("+block.Slot+") labelled '"+block.Label+"' updated successfully for '"+m[1]+"'.\",\"success\":true}")
							return
						} else {
							fmt.Fprintf(w, "{\"message\":\"Slot not found (blocked) for '"+m[1]+"'.\",\"success\":false}")
							return
						}
					} else {
						fmt.Fprintf(w, "{\"message\":\"Invalid value for slot.\",\"success\":false}")
						return
					}
				} else {
					fmt.Fprintf(w, "{\"message\":\"User doesnot exists.\",\"success\":false}")
					return
				}

			} else if m[2] == "scheduleMeeting" {
				var meet req_meeting
				reqBody, _ := ioutil.ReadAll(r.Body)

				err0 := json.Unmarshal(reqBody, &meet)
				if err0 != nil {
					fmt.Fprintf(w, "{\"message\":\"Please supply a json body. "+err0.Error()+"\",\"success\":false}")
					return
				}
				rows, err := db.Query(`SELECT * FROM user WHERE name = '` + m[1] + `';`)
				if err != nil {
					fmt.Fprintf(w, "%q\n", err)
					return
				} else if rows.Next() {
					var user1 user
					rows.Scan(&user1.Id, &user1.Name, &user1.Hod)
					rows.Close()
					if validDate.MatchString(meet.Slot) {
						rows3, err3 := db.Query(`SELECT label FROM meeting WHERE creator_id=` + fmt.Sprint(user1.Id) + ` AND user_id=` + fmt.Sprint(user1.Id) + ` AND datetime(slot) = datetime('` + meet.Slot + `')`)
						if err3 != nil {
							log.Printf("%q\n", err3)
							return
						} else if rows3.Next() {
							if meet.Label == "" {
								rows3.Scan(&meet.Label)
							} else {
								_, err4 := db.Exec(`UPDATE meeting SET label = '` + meet.Label + `' WHERE creator_id=` + fmt.Sprint(user1.Id) + ` AND datetime(slot)=datetime('` + meet.Slot + `')`)
								if err4 != nil {
									log.Printf("%q\n", err4)
									return
								}
							}
							rows3.Close()
							if meet.Users != nil {
								_, err2 := db.Exec(`DELETE FROM meeting WHERE creator_id=` + fmt.Sprint(user1.Id) + ` AND datetime(slot) = datetime('` + meet.Slot + `')`)
								if err2 != nil {
									log.Printf("%q\n", err2)
									return
								}
								var users []user
								for _, u := range meet.Users {
									rowsu, erru := db.Query(`SELECT * FROM user WHERE name='` + u + `'`)
									if erru != nil {
										log.Printf("%q\n", erru)
										return
									}
									if rowsu.Next() {
										var useru user
										rowsu.Scan(&useru.Id, &useru.Name, &useru.Hod)
										rowsu.Close()
										rowsu2, erru2 := db.Query(`SELECT * FROM meeting WHERE user_id=` + fmt.Sprint(useru.Id) + ` AND datetime(slot) = datetime('` + meet.Slot + `')`)
										if erru2 != nil {
											log.Printf("%q\n", erru2)
											return
										}
										if rowsu2.Next() {
											fmt.Fprintf(w, "{\"message\":\"Slot is already reserved (blocked) for '"+u+"'.\",\"success\":false}")
											return
										}
										rowsu2.Close()
										users = append(users, useru)
									} else {
										fmt.Fprintf(w, "{\"message\":\"User "+u+" doesnot exists.\",\"success\":false}")
										return
									}

								}
								_, err4 := db.Exec(`INSERT INTO meeting (creator_id,user_id,slot,label) VALUES(` + fmt.Sprint(user1.Id) + `,` + fmt.Sprint(user1.Id) + `,'` + meet.Slot + `','` + meet.Label + `')`)
								if err4 != nil {
									log.Printf("%q\n", err4)
									return
								} else {
									for _, useru := range users {
										_, erru := db.Exec(`INSERT INTO meeting (creator_id,user_id,slot,label) VALUES(` + fmt.Sprint(user1.Id) + `,` + fmt.Sprint(useru.Id) + `,'` + meet.Slot + `','` + meet.Label + `')`)
										if erru != nil {
											log.Printf("%q\n", erru)
											return
										}
									}
								}
							}
							fmt.Fprintf(w, "{\"message\":\"Meeting set at slot ("+meet.Slot+") labelled'"+meet.Label+"' successfully by '"+m[1]+"' with "+fmt.Sprint(meet.Users)+".\",\"success\":true}")
							return
						} else {
							fmt.Fprintf(w, "{\"message\":\"Slot (meeting) for '"+m[1]+"' not found.\",\"success\":false}")
							return
						}
					} else {
						fmt.Fprintf(w, "{\"message\":\"Invalid value for slot.\",\"success\":false}")
						return
					}
				} else {
					fmt.Fprintf(w, "{\"message\":\"User doesnot exists.\",\"success\":false}")
					return
				}

			} else {
				fmt.Fprintf(w, "Unknown Operation")
			}
		} else if r.Method == "DELETE" {
			if m[2] == "" {
				rows, err := db.Query(`SELECT * FROM user WHERE name = '` + m[1] + `';`)
				if err != nil {
					fmt.Fprintf(w, "%q\n", err)
					return
				} else if rows.Next() {
					rows.Close()
					_, err2 := db.Exec(`DELETE FROM user WHERE name="` + m[1] + `"`)
					if err2 != nil {
						log.Printf("%q\n", err2)
						return
					} else {
						fmt.Fprintf(w, "{\"message\":\"User '"+m[1]+"' deleted successfully.\",\"success\":true}")
						return
					}
				} else {
					fmt.Fprintf(w, "{\"message\":\"User doesnot exist.\",\"success\":false}")
					return
				}
			} else if m[2] == "blockCalendar" {
				var block req_blocked
				reqBody, _ := ioutil.ReadAll(r.Body)

				err0 := json.Unmarshal(reqBody, &block)
				if err0 != nil {
					fmt.Fprintf(w, "{\"message\":\"Please supply a json body. "+err0.Error()+"\",\"success\":false}")
					return
				}
				rows, err := db.Query(`SELECT * FROM user WHERE name = '` + m[1] + `';`)
				if err != nil {
					fmt.Fprintf(w, "%q\n", err)
					return
				} else if rows.Next() {
					var user1 user
					rows.Scan(&user1.Id, &user1.Name, &user1.Hod)
					rows.Close()
					if validDate.MatchString(block.Slot) {
						rows2, err2 := db.Query(`SELECT * FROM blocked WHERE user_id=` + fmt.Sprint(user1.Id) + ` AND datetime(slot) = datetime('` + block.Slot + `')`)
						if err2 != nil {
							log.Printf("%q\n", err2)
							return
						}
						if rows2.Next() {
							rows2.Close()
							_, err4 := db.Exec(`DELETE FROM blocked WHERE user_id=` + fmt.Sprint(user1.Id) + ` AND datetime(slot)=datetime('` + block.Slot + `')`)
							if err4 != nil {
								log.Printf("%q\n", err4)
								return
							}
							fmt.Fprintf(w, "{\"message\":\"Blocked slot ("+block.Slot+") was unblocked successfully for '"+m[1]+"'.\",\"success\":true}")
							return
						} else {
							fmt.Fprintf(w, "{\"message\":\"Slot not found (blocked) for '"+m[1]+"'.\",\"success\":false}")
							return
						}
					} else {
						fmt.Fprintf(w, "{\"message\":\"Invalid value for slot.\",\"success\":false}")
						return
					}
				} else {
					fmt.Fprintf(w, "{\"message\":\"User doesnot exists.\",\"success\":false}")
					return
				}

			} else if m[2] == "scheduleMeeting" {
				var meet req_meeting
				reqBody, _ := ioutil.ReadAll(r.Body)

				err0 := json.Unmarshal(reqBody, &meet)
				if err0 != nil {
					fmt.Fprintf(w, "{\"message\":\"Please supply a json body. "+err0.Error()+"\",\"success\":false}")
					return
				}
				rows, err := db.Query(`SELECT * FROM user WHERE name = '` + m[1] + `';`)
				if err != nil {
					fmt.Fprintf(w, "%q\n", err)
					return
				} else if rows.Next() {
					var user1 user
					rows.Scan(&user1.Id, &user1.Name, &user1.Hod)
					rows.Close()
					if validDate.MatchString(meet.Slot) {
						rows3, err3 := db.Query(`SELECT label FROM meeting WHERE creator_id=` + fmt.Sprint(user1.Id) + ` AND user_id=` + fmt.Sprint(user1.Id) + ` AND datetime(slot) = datetime('` + meet.Slot + `')`)
						if err3 != nil {
							log.Printf("%q\n", err3)
							return
						} else if rows3.Next() {
							rows3.Close()
							_, err2 := db.Exec(`DELETE FROM meeting WHERE creator_id=` + fmt.Sprint(user1.Id) + ` AND datetime(slot) = datetime('` + meet.Slot + `')`)
							if err2 != nil {
								log.Printf("%q\n", err2)
								return
							}
							fmt.Fprintf(w, "{\"message\":\"Meeting set at slot ("+meet.Slot+") by '"+m[1]+"' was deleted successfully.\",\"success\":true}")
							return
						} else {
							fmt.Fprintf(w, "{\"message\":\"Slot (meeting) for '"+m[1]+"' not found.\",\"success\":false}")
							return
						}
					} else {
						fmt.Fprintf(w, "{\"message\":\"Invalid value for slot.\",\"success\":false}")
						return
					}
				} else {
					fmt.Fprintf(w, "{\"message\":\"User doesnot exists.\",\"success\":false}")
					return
				}

			} else {
				fmt.Fprintf(w, "Unknown Operation")
			}
		} else {
			fmt.Fprintf(w, "Unknown Method. Allowed [GET,POST,PUT,PATCH,DELETE]. ")
		}
	})
	fmt.Println("Started listening to 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
