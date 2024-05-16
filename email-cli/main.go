package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
)

var tokenFile = "tokens.json"

var rootCmd = &cobra.Command{
	Use:   "email-cli",
	Short: "CLI for managing emails",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(getEmailsCmd)
	rootCmd.AddCommand(getEmailByIDCmd)
	rootCmd.AddCommand(deleteEmailByIDCmd)
	rootCmd.AddCommand(deleteAllUserEmailsCmd)
	rootCmd.AddCommand(updateEmailsForUserCmd)
	rootCmd.AddCommand(getAllUsersCmd)
	rootCmd.AddCommand(deleteUserCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login with Google and get tokens",
	Run: func(cmd *cobra.Command, args []string) {
		var wg sync.WaitGroup
		wg.Add(1)

		// Открытие браузера для логина
		url := "http://localhost:8080/api/auth/google_login"
		fmt.Println("Opening browser for login, please complete the login process.")
		openBrowser(url)

		// Считывание токенов из консоли
		fmt.Println("Please enter the tokens in JSON format:")
		tokens := make(map[string]string)
		var input string
		fmt.Scanln(&input)
		if err := json.Unmarshal([]byte(input), &tokens); err != nil {
			fmt.Println("Error decoding input:", err)
			return
		}

		file, err := os.Create(tokenFile)
		if err != nil {
			fmt.Println("Error creating token file:", err)
			return
		}
		defer file.Close()

		json.NewEncoder(file).Encode(tokens)
		fmt.Println("Tokens saved successfully")
		wg.Done()
		wg.Wait()
	},
}

func openBrowser(url string) {
	var err error
	switch os := runtime.GOOS; os {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		fmt.Println("Failed to open browser:", err)
	}
}

func readTokens() (map[string]string, error) {
	file, err := os.Open(tokenFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var tokens map[string]string
	if err := json.NewDecoder(file).Decode(&tokens); err != nil {
		return nil, err
	}
	return tokens, nil
}

func getAccessToken() (string, error) {
	tokens, err := readTokens()
	if err != nil {
		return "", err
	}
	return tokens["access_token"], nil
}

var getEmailsCmd = &cobra.Command{
	Use:   "get-emails [userID]",
	Short: "Get list of emails for a user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		userID := args[0]
		token, err := getAccessToken()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8080/api/emails/user/%s", userID), nil)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		var emails []map[string]interface{}
		if err := json.Unmarshal(body, &emails); err != nil {
			fmt.Println("Error parsing response:", err)
			return
		}

		// Создаем папку htmls, если она не существует
		if err := os.MkdirAll("htmls", os.ModePerm); err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}

		for i, email := range emails {
			subject := email["subject"].(string)
			body := email["body"].(string)
			filename := filepath.Join("htmls", fmt.Sprintf("%d.html", i+1))
			if err := ioutil.WriteFile(filename, []byte(body), 0644); err != nil {
				fmt.Println("Error writing file:", err)
				return
			}

			// Сохраняем метаинформацию о письме
			metaFile := filepath.Join("htmls", fmt.Sprintf("%d.json", i+1))
			metaInfo := map[string]string{"subject": subject}
			if metaData, err := json.Marshal(metaInfo); err == nil {
				if err := ioutil.WriteFile(metaFile, metaData, 0644); err != nil {
					fmt.Println("Error writing meta file:", err)
					return
				}
			} else {
				fmt.Println("Error marshaling meta info:", err)
				return
			}

			openBrowser("file://" + filename)
		}
	},
}

var getEmailByIDCmd = &cobra.Command{
	Use:   "get-email [userID] [emailID]",
	Short: "Get email by ID for a user",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		userID, emailID := args[0], args[1]
		token, err := getAccessToken()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8080/api/emails/user/%s/%s", userID, emailID), nil)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		var email map[string]interface{}
		if err := json.Unmarshal(body, &email); err != nil {
			fmt.Println("Error parsing response:", err)
			return
		}

		// Создаем папку htmls, если она не существует
		if err := os.MkdirAll("htmls", os.ModePerm); err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}

		filename := filepath.Join("htmls", fmt.Sprintf("%s.html", emailID))
		if err := ioutil.WriteFile(filename, []byte(email["body"].(string)), 0644); err != nil {
			fmt.Println("Error writing file:", err)
			return
		}

		// Сохраняем метаинформацию о письме
		metaFile := filepath.Join("htmls", fmt.Sprintf("%s.json", emailID))
		metaInfo := map[string]string{"subject": email["subject"].(string)}
		if metaData, err := json.Marshal(metaInfo); err == nil {
			if err := ioutil.WriteFile(metaFile, metaData, 0644); err != nil {
				fmt.Println("Error writing meta file:", err)
				return
			}
		} else {
			fmt.Println("Error marshaling meta info:", err)
			return
		}

		openBrowser("file://" + filename)
	},
}

var deleteEmailByIDCmd = &cobra.Command{
	Use:   "delete-email [userID] [emailID]",
	Short: "Delete email by ID for a user",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		userID, emailID := args[0], args[1]
		token, err := getAccessToken()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		req, err := http.NewRequest("DELETE", fmt.Sprintf("http://localhost:8080/api/emails/user/%s/%s", userID, emailID), nil)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer resp.Body.Close()

		fmt.Println("Email deleted successfully")
	},
}

var deleteAllUserEmailsCmd = &cobra.Command{
	Use:   "delete-all-emails [userID]",
	Short: "Delete all emails for a user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		userID := args[0]
		token, err := getAccessToken()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		req, err := http.NewRequest("DELETE", fmt.Sprintf("http://localhost:8080/api/emails/user/%s", userID), nil)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer resp.Body.Close()

		fmt.Println("All emails deleted successfully")
	},
}

var updateEmailsForUserCmd = &cobra.Command{
	Use:   "update-emails [userID]",
	Short: "Update emails for a user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		userID := args[0]
		token, err := getAccessToken()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:8080/api/emails/user/%s/update", userID), nil)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer resp.Body.Close()

		fmt.Println("Emails updated successfully")
	},
}

var getAllUsersCmd = &cobra.Command{
	Use:   "get-users",
	Short: "Get list of all users",
	Run: func(cmd *cobra.Command, args []string) {
		token, err := getAccessToken()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		req, err := http.NewRequest("GET", "http://localhost:8080/api/users", nil)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
	},
}

var deleteUserCmd = &cobra.Command{
	Use:   "delete-user [userID]",
	Short: "Delete a user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		userID := args[0]
		token, err := getAccessToken()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		req, err := http.NewRequest("DELETE", fmt.Sprintf("http://localhost:8080/api/users/%s", userID), nil)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer resp.Body.Close()

		fmt.Println("User deleted successfully")
	},
}
