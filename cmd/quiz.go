package cmd

// Importera nödvändiga paket
import (
    "bufio"             // Paket för att läsa inmatning från användaren rad för rad
    "bytes"             // Paket för att hantera bytebuffertar
    "encoding/json"     // Paket för JSON-kodning och avkodning
    "fmt"               // Paket för formaterad utskrift
    "net/http"          // Paket för HTTP-kommunikation
    "os"                // Paket för operativsystemsfunktioner
    "strconv"           // Paket för att konvertera strängar till andra typer
    "strings"           // Paket för stränghantering
    "github.com/spf13/cobra" // Externt paket för att skapa CLI-applikationer
)

// Definition av struktur för frågor
type Question struct {
    ID         string   `json:"id"`        // ID för frågan
    Query      string   `json:"query"`     // Själva frågan
    Options    []string `json:"options"`   // Alternativ för svar
    CorrectAns int      `json:"correctAns"` // Index för det korrekta svaret i alternativen
}

// Definition av struktur för svar
type Submission struct {
    Answers map[string]int `json:"answers"` // Kartläggning av fråge-ID till användarens svar
}

// Skapa ett nytt Cobra-kommando för quizen
var quizCmd = &cobra.Command{
    Use:   "quiz",                                 // Användningstext för kommandot
    Short: "Starts the quiz and interacts with the server", // Kort beskrivning av kommandot
    Run: func(cmd *cobra.Command, args []string) { // Funktion som körs när kommandot anropas
        getQuestionsAndCollectAnswers() // Anropa funktionen för att hämta och samla in svar
    },
}

// Funktion för att lägga till kommandot i init-funktionen för rootCmd
func init() {
    rootCmd.AddCommand(quizCmd) // Lägg till quizCmd som ett underkommando till rootCmd
}

// Funktion för att hämta frågor från servern och samla in användarens svar
func getQuestionsAndCollectAnswers() {
    // Gör en HTTP GET-förfrågan till servern för att hämta frågor
    resp, err := http.Get("http://localhost:8080/question")
    if err != nil {
        fmt.Println("Could not fetch questions:", err) // Hantera fel om frågorna inte kan hämtas
        return
    }
    defer resp.Body.Close() // Stäng svarskroppen när funktionen är klar

    var questions []Question // Skapa en tom lista för frågor
    // Avkoda JSON-svaret till frågelistan
    if err := json.NewDecoder(resp.Body).Decode(&questions); err != nil {
        fmt.Println("Could not parse questions:", err) // Hantera fel om frågorna inte kan tolkas
        return
    }

    // Skapa en karta för att lagra användarens svar
    answers := make(map[string]int)
    scanner := bufio.NewScanner(os.Stdin) // Skapa en scanner för att läsa inmatning från användaren

    // Iterera över varje fråga och låt användaren svara
    for _, q := range questions {
        fmt.Printf("Question: %s\n", q.Query) // Skriv ut frågetexten för användaren
        for i, option := range q.Options {
            fmt.Printf("%d. %s\n", i+1, option) // Skriv ut varje svarsalternativ
        }
        fmt.Print("Enter the number of your answer: ") // Be användaren ange sitt svar
        scanner.Scan()                                   // Läs in användarens svar från terminalen
        input := scanner.Text()                         // Hämta användarens svar som en sträng
        answer, err := strconv.Atoi(strings.TrimSpace(input)) // Konvertera svaret till en integer
        if err != nil {
            fmt.Println("Invalid input, please enter a number. Error:", err) // Hantera ogiltigt svar från användaren
            continue // Gå till nästa fråga om det finns ett felaktigt svar
        }
        answers[q.ID] = answer - 1 // Lagra användarens svar i kartan, justera index för att matcha nollbaserat index
    }

    submitAnswers(answers) // Skicka användarens svar till servern för att beräkna poäng
}

// Funktion för att skicka användarens svar till servern för beräkning av poäng
func submitAnswers(answers map[string]int) {
    // Kodera användarens svar som JSON-data
    jsonData, err := json.Marshal(Submission{Answers: answers})
    if err != nil {
        fmt.Println("Error preparing request:", err) // Hantera fel om JSON-data inte kan skapas
        return
    }

    // Skicka JSON-data till servern via en HTTP POST-förfrågan
    resp, err := http.Post("http://localhost:8080/submit", "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        fmt.Println("Error submitting answers:", err) // Hantera fel om svaren inte kan skickas till servern
        return
    }
    defer resp.Body.Close() // Stäng svarskroppen när funktionen är klar

    // Avkoda svaret från servern till en struktur för att hämta användarens poäng
    var result struct {
        Score int `json:"score"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        fmt.Println("Error reading response:", err) // Hantera fel om svaret inte kan tolkas
        return
    }

    fmt.Printf("You scored %d points!\n", result.Score) // Skriv ut användarens poäng till terminalen
}
