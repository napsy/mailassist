package main

import (
	"flag"
	"log"
	"os"
	"time"
)

func main() {
	var (
		modelFlag = flag.String("model", "zephyr", "llm model (e.g. mistral, gpt-4, ...)")
		llmFlag   = flag.String("llm", "ollama", "choose from openai or ollama")
		tokenFlag = flag.String("token", "XYZ", "some llm require tokem authentication")

		ai  LLM
		err error
	)

	flag.Parse()

	switch *llmFlag {
	case "ollama":
		ai, err = newOllama(*modelFlag)
		if err != nil {
			log.Fatalf("Could not initialize AI: %v", err)
		}
	case "openai":
		*tokenFlag, _ = os.LookupEnv("OPENAI_KEY")
		ai, err = newOpenAI(*tokenFlag)
		if err != nil {
			log.Fatalf("Could not initialize AI: %v", err)
		}
	}

	// TODO: insert your own bio
	ai.bio("I am Luka Napotnik, the VP of Engineering, my primary focus is team output and influence, product quality, infrastructure cost, retention and people growth.")

	gmail, err := newGmailProvider("credentials.json", 40)
	if err != nil {
		log.Fatalf("Could not initialize Gmail: %v", err)
	}

	d := newDesktop()
	web := newWebAPI()
	db := newLocalDB()

	mbox := newMailbox(gmail, ai)

	for {
		if err := mbox.fetch(); err != nil {
			log.Fatalf("Error fetching mail: %v", err)
		}
		mbox.summarize(func(from string, date string, subject string, message string, original string) {
			h := hashMail(date, from, subject)
			if db.wasRead(h) {
				return
			}
			db.markRead(h)
			d.notify(subject)
			web.push(from, date, subject, highlightPriority(markdownMessage(message)), markdownMessage(original))
		})
		time.Sleep(10 * time.Minute)
	}
}
