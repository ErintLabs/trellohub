package trello

import (
  . "github.com/ErintLabs/trellohub/genapi"
  "github.com/ErintLabs/trellohub/github"
  "net/url"
  "log"
)

// TODO: handle error responces from Trello

/* TODO comments */
type ListRef struct {
  ReposId   string    `json:"repos"`
  InboxId   string    `json:"inbox"`
  InWorksId string    `json:"works"`
  BlockedId string    `json:"block"`
  ReviewId  string    `json:"review"`
  MergedId  string    `json:"merged"`
  DeployId  string    `json:"deploy"`
  TestId    string    `json:"tested"`
  AcceptId  string    `json:"accept"`
}

type Payload struct {
  Action      struct {
    Type      string        `json:"type"`
    Data      struct {
      Member  string        `json:"idMember"`
      List    Object        `json:"list"`
      ChList  Checklist     `json:"checklist"`
      ChItem  CheckItem     `json:"checkItem"`
      Card    Card          `json:"card"`
      Old     struct {
        Name  string        `json:"name"`
        Desc  string        `json:"desc"`
      }                     `json:"old"`
      ListB   Object        `json:"listBefore"`
      ListA   Object        `json:"listAfter"`
      Attach  struct {
        URL   string        `json:"url"`
      }                     `json:"attachment"`
    }                       `json:"data"`
  }                         `json:"action"`
}

type Object struct {
  Id      string    `json:"id"`
  Name    string    `json:"name"`
}

/* TODO make some fields private */
type Trello struct {
  Token string
  Key string
  BoardId string
  Lists ListRef
  github *github.GitHub

  /* RenameThese to make sense */
  labelCache map[string]string
  userIdbyName map[string]string
  userNamebyId map[string]string

  cardById      map[string]*Card
  cardByIssue   map[string]*Card
}

func New(key string, token string, boardid string) *Trello {
  t := new(Trello)
  t.Token = token
  t.Key = key

  t.BoardId = t.getFullBoardId(boardid)

  return t
}

func (trello *Trello) Startup(github *github.GitHub) {
  trello.github = github

  trello.labelCache = make(map[string]string)
  trello.makeLabelCache()

  /* Note: we assume users don't change anyway so we only do trello at startup */
  trello.userIdbyName = make(map[string]string)
  trello.makeUserCache()

  trello.cardById = make(map[string]*Card)
  trello.cardByIssue = make(map[string]*Card)
  trello.makeCardCache()
}

func (trello *Trello) AuthQuery() string {
  return "key=" + trello.Key + "&token=" + trello.Token
}

func (trello *Trello) BaseURL() string {
  return "https://api.trello.com/1"
}

func (trello *Trello) getFullBoardId(boardid string) string {
  data := Object{}
  GenGET(trello, "/boards/" + boardid, &data)
  return data.Id
}

type webhookInfo struct {
  Id    string    `json:"id"`
  Model string    `json:"idModel"`
  URL   string    `json:"callbackURL"`
}

/* Checks that a webhook is installed over the board, in case it isn't creates one */
func (trello *Trello) EnsureHook(callbackURL string) {
  /* Check if we have a hook already */
  var data []webhookInfo
  GenGET(trello, "/token/" + trello.Token + "/webhooks/", &data)
  found := false

  for _, v := range data {
    /* Check if we have a hook for our own URL at same model */
    if v.Model == trello.BoardId {
      if v.URL == callbackURL {
        log.Print("Hook found, nothing to do here.")
        found = true
        break
      }
    }
  }

  /* If not, install one */
  if !found {
    /* TODO: save hook reference and uninstall maybe? */
    GenPOSTForm(trello, "/webhooks/", nil, url.Values{
      "name": { "trellohub for " + trello.BoardId },
      "idModel": { trello.BoardId },
      "callbackURL": { callbackURL } })

    log.Print("Webhook installed.")
  } else {
    log.Print("Reusing existing webhook.")
  }
}
