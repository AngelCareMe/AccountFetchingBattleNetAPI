package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"guildtracker/models"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	clientID     = ""
	clientSecret = ""
	redirectURI  = "http://localhost:8080/callback"
)

// Character представляет персонажа из API Blizzard
type Character struct {
	Name  string `json:"name"`
	ID    int    `json:"id"`
	Realm struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
	} `json:"realm"`
	Level int `json:"level"`
	Race  struct {
		Name string `json:"name"`
	} `json:"playable_race"`
	CharacterClass struct {
		Name string `json:"name"`
	} `json:"playable_class"`
}

// AccountProfile представляет профиль аккаунта из API Blizzard
type AccountProfile struct {
	WoWAccounts []struct {
		Characters []Character `json:"characters"`
	} `json:"wow_accounts"`
}

func init() {
	log.Println("Using hardcoded credentials")
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./templates/login.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, fmt.Sprintf("Failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	state := uuid.New().String()
	http.SetCookie(w, &http.Cookie{
		Name:   "oauth_state",
		Value:  state,
		Path:   "/",
		MaxAge: 300,
	})

	authURL := fmt.Sprintf(
		"https://eu.battle.net/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=wow.profile&state=%s",
		clientID, url.QueryEscape(redirectURI), state,
	)
	log.Printf("Redirecting to: %s", authURL)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// getAccessToken получает токен доступа по коду авторизации
func getAccessToken(code string) (string, error) {
	tokenURL := "https://eu.battle.net/oauth/token"
	formData := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
	}

	resp, err := http.PostForm(tokenURL, formData)
	if err != nil {
		return "", fmt.Errorf("failed to get token: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read token response: %v", err)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token: %v", err)
	}

	log.Printf("Access Token: %s", tokenResp.AccessToken)
	return tokenResp.AccessToken, nil
}

// fetchAccountProfile запрашивает профиль аккаунта
func fetchAccountProfile(client *http.Client, accessToken string) (AccountProfile, error) {
	profileURL := fmt.Sprintf("https://eu.api.blizzard.com/profile/user/wow?namespace=profile-eu&locale=ru_RU&access_token=%s", accessToken)
	req, err := http.NewRequest("GET", profileURL, nil)
	if err != nil {
		return AccountProfile{}, fmt.Errorf("failed to create profile request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("User-Agent", "Go-Client/1.0")

	profileResp, err := client.Do(req)
	if err != nil {
		return AccountProfile{}, fmt.Errorf("failed to get profile: %v", err)
	}
	defer profileResp.Body.Close()

	profileBody, err := io.ReadAll(profileResp.Body)
	if err != nil {
		return AccountProfile{}, fmt.Errorf("failed to read profile response: %v", err)
	}

	if profileResp.StatusCode != http.StatusOK {
		time.Sleep(1 * time.Second)
		profileResp, err = client.Do(req)
		if err != nil {
			return AccountProfile{}, fmt.Errorf("failed to get profile on retry: %v", err)
		}
		defer profileResp.Body.Close()

		profileBody, err = io.ReadAll(profileResp.Body)
		if err != nil {
			return AccountProfile{}, fmt.Errorf("failed to read profile response on retry: %v", err)
		}
		if profileResp.StatusCode != http.StatusOK {
			return AccountProfile{}, fmt.Errorf("profile request failed with status: %d, body: %s", profileResp.StatusCode, string(profileBody))
		}
	}

	var accountProfile AccountProfile
	if err := json.Unmarshal(profileBody, &accountProfile); err != nil {
		return AccountProfile{}, fmt.Errorf("failed to decode profile: %v", err)
	}

	return accountProfile, nil
}

// fetchCharacterDetails запрашивает дополнительные данные о персонаже (Ilvl, Guild, MythicRating)
func fetchCharacterDetails(client *http.Client, character Character, accessToken string) models.Character {
	characterData := models.Character{
		Name:         character.Name,
		Class:        character.CharacterClass.Name,
		Level:        character.Level,
		Race:         character.Race.Name,
		Ilvl:         0,
		Guild:        "",
		MythicRating: 0.0,
	}

	characterNameLower := strings.ToLower(character.Name)

	// Запрашиваем экипировку для Ilvl
	equipmentURL := fmt.Sprintf("https://eu.api.blizzard.com/profile/wow/character/%s/%s/equipment?namespace=profile-eu&locale=ru_RU&access_token=%s",
		character.Realm.Slug, url.PathEscape(characterNameLower), accessToken)
	eqReq, err := http.NewRequest("GET", equipmentURL, nil)
	if err != nil {
		log.Printf("Failed to create equipment request for %s: %v", character.Name, err)
	} else {
		eqReq.Header.Set("Authorization", "Bearer "+accessToken)
		eqReq.Header.Set("User-Agent", "Go-Client/1.0")
		eqResp, err := client.Do(eqReq)
		if err != nil {
			log.Printf("Failed to get equipment for %s: %v", character.Name, err)
		} else {
			defer eqResp.Body.Close()
			eqBody, err := io.ReadAll(eqResp.Body)
			if err != nil {
				log.Printf("Failed to read equipment response for %s: %v", character.Name, err)
			} else if eqResp.StatusCode != http.StatusOK {
				log.Printf("Equipment request for %s failed with status: %d, body: %s", character.Name, eqResp.StatusCode, string(eqBody))
			} else {
				var equipment struct {
					EquippedItems []struct {
						ItemLevel struct {
							Value int `json:"value"`
						} `json:"level"`
					} `json:"equipped_items"`
				}
				if err := json.Unmarshal(eqBody, &equipment); err != nil {
					log.Printf("Failed to decode equipment for %s: %v", character.Name, err)
				} else {
					var totalIlvl int
					if len(equipment.EquippedItems) > 0 {
						for _, item := range equipment.EquippedItems {
							totalIlvl += item.ItemLevel.Value
						}
						characterData.Ilvl = totalIlvl / len(equipment.EquippedItems)
					}
				}
			}
		}
	}

	time.Sleep(100 * time.Millisecond)

	// Запрашиваем профиль персонажа для гильдии
	charURL := fmt.Sprintf("https://eu.api.blizzard.com/profile/wow/character/%s/%s?namespace=profile-eu&locale=ru_RU&access_token=%s",
		character.Realm.Slug, url.PathEscape(characterNameLower), accessToken)
	charReq, err := http.NewRequest("GET", charURL, nil)
	if err != nil {
		log.Printf("Failed to create character request for %s: %v", character.Name, err)
	} else {
		charReq.Header.Set("Authorization", "Bearer "+accessToken)
		charReq.Header.Set("User-Agent", "Go-Client/1.0")
		charResp, err := client.Do(charReq)
		if err != nil {
			log.Printf("Failed to get character profile for %s: %v", character.Name, err)
		} else {
			defer charResp.Body.Close()
			charBody, err := io.ReadAll(charResp.Body)
			if err != nil {
				log.Printf("Failed to read character profile for %s: %v", character.Name, err)
			} else if charResp.StatusCode != http.StatusOK {
				log.Printf("Character profile request for %s failed with status: %d, body: %s", character.Name, charResp.StatusCode, string(charBody))
			} else {
				var charProfile struct {
					Guild struct {
						Name string `json:"name"`
					} `json:"guild"`
				}
				if err := json.Unmarshal(charBody, &charProfile); err != nil {
					log.Printf("Failed to decode character profile for %s: %v", character.Name, err)
				} else {
					characterData.Guild = charProfile.Guild.Name
				}
			}
		}
	}

	time.Sleep(100 * time.Millisecond)

	// Запрашиваем M+ рейтинг
	mplusURL := fmt.Sprintf("https://eu.api.blizzard.com/profile/wow/character/%s/%s/mythic-keystone-profile?namespace=profile-eu&locale=ru_RU&access_token=%s",
		character.Realm.Slug, url.PathEscape(characterNameLower), accessToken)
	mplusReq, err := http.NewRequest("GET", mplusURL, nil)
	if err != nil {
		log.Printf("Failed to create M+ request for %s: %v", character.Name, err)
	} else {
		mplusReq.Header.Set("Authorization", "Bearer "+accessToken)
		mplusReq.Header.Set("User-Agent", "Go-Client/1.0")
		mplusResp, err := client.Do(mplusReq)
		if err != nil {
			log.Printf("Failed to get M+ profile for %s: %v", character.Name, err)
		} else {
			defer mplusResp.Body.Close()
			mplusBody, err := io.ReadAll(mplusResp.Body)
			if err != nil {
				log.Printf("Failed to read M+ profile for %s: %v", character.Name, err)
			} else if mplusResp.StatusCode != http.StatusOK {
				log.Printf("M+ profile request for %s failed with status: %d, body: %s", character.Name, mplusResp.StatusCode, string(mplusBody))
			} else {
				var mplusProfile struct {
					CurrentMythicRating struct {
						Rating float64 `json:"rating"`
					} `json:"current_mythic_rating"`
				}
				if err := json.Unmarshal(mplusBody, &mplusProfile); err != nil {
					log.Printf("Failed to decode M+ profile for %s: %v", character.Name, err)
				} else {
					characterData.MythicRating = mplusProfile.CurrentMythicRating.Rating
				}
			}
		}
	}

	return characterData
}

// renderTemplate рендерит шаблон с данными персонажей
func renderTemplate(w http.ResponseWriter, characters []models.Character) error {
	tmpl, err := template.ParseFiles("./templates/characters.html")
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	data := struct {
		Characters []models.Character
	}{
		Characters: characters,
	}
	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}
	return nil
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем state
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		http.Error(w, "State cookie missing", http.StatusBadRequest)
		return
	}
	state := r.URL.Query().Get("state")
	if state == "" || state != stateCookie.Value {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "code is required", http.StatusBadRequest)
		return
	}

	// Получаем токен доступа
	accessToken, err := getAccessToken(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	client := &http.Client{}

	// Получаем профиль аккаунта
	accountProfile, err := fetchAccountProfile(client, accessToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Собираем список персонажей
	var characters []models.Character
	for _, account := range accountProfile.WoWAccounts {
		for _, character := range account.Characters {
			characterData := fetchCharacterDetails(client, character, accessToken)
			characters = append(characters, characterData)
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Отладка: проверяем, что данные собраны
	log.Printf("Characters to render: %+v", characters)

	// Рендерим шаблон
	if err := renderTemplate(w, characters); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
