package main

// Importera nödvändiga paket
import (
	"encoding/json" // Paket för JSON-kodning och avkodning
	"fmt"           // Paket för formaterad utskrift
	"log"           // Paket för att logga fel
	"net/http"      // Paket för HTTP-kommunikation
	"sync"          // Paket för synkronisering av trådar
)

// Definition av struktur för frågor
type Question struct {
	ID         string   `json:"id"`        // ID för frågan
	Query      string   `json:"query"`     // Själva frågan
	Options    []string `json:"options"`   // Alternativ för svar
	CorrectAns int      `json:"correctAns"` // Index för det korrekta svaret i alternativen
}

// Definition av struktur för quiz
type Quiz struct {
	Questions []Question `json:"questions"` // Lista med frågor
	Current   int         // Aktuell fråga
	Score     int         // Poäng
	Finished  bool        // Indikerar om quizen är avslutad eller inte
	mux       sync.Mutex  // Mutex för att synkronisera åtkomst till delade data
}

// Funktion för att skapa en ny quiz med fördefinierade frågor
func NewQuiz() *Quiz {
	return &Quiz{
		Questions: []Question{ // Fördefinierade frågor
			{ID: "1", Query: "Vad är det enda landet som börjar på Q?", Options: []string{"Quatar", "Quebec", "Qatar"}, CorrectAns: 2},
			{ID: "2", Query: "Vilken färg får du om du blandar rött och vitt?", Options: []string{"Lila", "Orange", "Rosa"}, CorrectAns: 2},
			{ID: "3", Query: "Vilket djur är känt för att vara det snabbaste djuret i världen på land?", Options: []string{"Geparden", "Lejonet", "Hästen"}, CorrectAns: 0},
			{ID: "4", Query: "Vad heter huvudstaden i Australien?", Options: []string{"Sydney", "Canberra", "Melbourne"}, CorrectAns: 1},
			{ID: "5", Query: "Vilken planet är känd som 'Röda planeten'?", Options: []string{"Jupiter", "Mars", "Venus"}, CorrectAns: 1},
			{ID: "6", Query: "Vilket år landade den första människan på månen?", Options: []string{"1969", "1972", "1967"}, CorrectAns: 0},
			{ID: "7", Query: "Vilken är den enda maten som aldrig blir dålig?", Options: []string{"Ris", "Honung", "Torkad pasta"}, CorrectAns: 1},
			{ID: "8", Query: "Vilken kroppsdelen fortsätter växa under hela ditt liv?", Options: []string{"Fingrar", "Näsan och öronen", "Fötter"}, CorrectAns: 1},
			{ID: "9", Query: "Vilken är den mest spelade låten på Spotify?", Options: []string{"'Shape of You' av Ed Sheeran", "'Blinding Lights' av The Weeknd", "'Despacito' av Luis Fonsi"}, CorrectAns: 0},
			{ID: "10", Query: "Vad är det som går och går men aldrig kommer till dörren?", Options: []string{"Tiden", "Floden", "Klockan"}, CorrectAns: 2},
		},
		Current:   0,      // Sätter aktuell fråga till första frågan
		Score:     0,      // Nollställer poängen
		Finished:  false,  // Indikerar att quizen inte är avslutad vid start
	}
}

// Funktion för att servera nästa fråga eller indikera att quizen är avslutad
func (q *Quiz) serveQuestion(w http.ResponseWriter, r *http.Request) {
	q.mux.Lock()          // Lås åtkomsten till quizen för att undvika datakonflikter
	defer q.mux.Unlock()  // Se till att låset frigörs efter att funktionen är klar

	if q.Finished { // Om quizen är avslutad
		http.Error(w, "Quiz completed", http.StatusForbidden) // Skicka ett felmeddelande till klienten
		return
	}

	question := q.Questions[q.Current] // Hämta aktuell fråga från quizen
	q.Current++                         // Öka index för aktuell fråga för att hämta nästa fråga vid nästa förfrågan

	if q.Current >= len(q.Questions) { // Om det var den sista frågan
		q.Finished = true // Markera quizen som avslutad
	}

	w.Header().Set("Content-Type", "application/json") // Sätt rätt HTTP-headers för svar
	json.NewEncoder(w).Encode(question)                // Kodera frågan som JSON och skicka till klienten
}

// Funktion för att hantera användarens svar och beräkna poäng
func (q *Quiz) handleSubmission(w http.ResponseWriter, r *http.Request) {
	var submission struct {
		Answer int `json:"answer"` // Struktur för användarens svar
	}

	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil { // Avkoda användarens svar från JSON
		http.Error(w, "Invalid input", http.StatusBadRequest) // Hantera felaktigt användarinput
		return
	}

	correctAnswer := q.Questions[q.Current-1].CorrectAns // Hämta index för det korrekta svaret
	if submission.Answer == correctAnswer {               // Om användarens svar är korrekt
		q.Score++ // Öka poängen
	}

	if q.Finished { // Om quizen är avslutad
		result := struct { // Skapa ett JSON-svar med användarens poäng
			Score int `json:"score"` // Poängen
		}{
			Score: q.Score, // Användarens totala poäng
		}
		json.NewEncoder(w).Encode(result) // Skicka JSON-svaret till klienten
	} else {
		q.serveQuestion(w, r) // Servera nästa fråga direkt om quizen inte är avslutad
	}
}

// Huvudfunktionen för att starta servern och hantera förfrågningar
func main() {
	quiz := NewQuiz() // Skapa en ny quiz
	http.HandleFunc("/question", quiz.serveQuestion) // Sätt upp en HTTP-handlare för att hantera frågor
	http.HandleFunc("/submit", quiz.handleSubmission) // Sätt upp en HTTP-handlare för att hantera svar

	fmt.Println("Server starting on port :8080...") // Skriv ut meddelande när servern startar
	log.Fatal(http.ListenAndServe(":8080", nil))    // Starta servern och logga eventuella fel
}
