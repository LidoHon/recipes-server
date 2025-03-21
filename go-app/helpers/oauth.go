package helpers

import (
	"context"
	"fmt"
	"log"

	"github.com/LidoHon/recipes-server/libs"
	"github.com/shurcooL/graphql"
)

func HandleAuth(email, userName, profile, providerId, providerName string) (token string, refreshToken string, id int, role string, err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := libs.SetupGraphqlClient()

	// Query existing user
	var query struct {
		User []struct {
			ID       graphql.Int    `graphql:"id"`
			UserName graphql.String `graphql:"username"`
			Email    graphql.String `graphql:"email"`
			Password graphql.String `graphql:"password"`
			Role     graphql.String `graphql:"role"`
			TokenId  graphql.String `graphql:"tokenId"`
			GoogleID graphql.String `graphql:"google_id"`
			GithubID graphql.String `graphql:"github_id"`
		} `graphql:"users(where: {email: {_eq: $email}})"`
	}

	queryVars := map[string]interface{}{
		"email": graphql.String(email),
	}

	err = client.Query(ctx, &query, queryVars)
	if err != nil {
		log.Println("Cpundn't query the user", err.Error())
		return "", "", 0, "", err
	}

	// If user does not exist, create a new one
	if len(query.User) == 0 {
		var mutation struct {
			InsertUser struct {
				ID       graphql.Int    `graphql:"id"`
				UserName graphql.String `graphql:"username"`
				Email    graphql.String `graphql:"email"`
				Profile  graphql.String `graphql:"profile"`
				Role     graphql.String `graphql:"role"`
				TokenId  graphql.String `graphql:"tokenId"`
				GoogleID graphql.String `graphql:"google_id"`
				GithubID graphql.String `graphql:"github_id"`
			} `graphql:"insert_users_one(object: {username: $userName, email: $email, profile: $profile, role: $role, is_email_verified: $is_email_verified, google_id: $googleID, github_id: $githubID})"`
		}

		var googleID, githubID graphql.String
		if providerName == "google" {
			googleID = graphql.String(providerId)
		} else if providerName == "github" {
			githubID = graphql.String(providerId)
		}

		mutationVars := map[string]interface{}{
			"userName":          graphql.String(userName),
			"email":             graphql.String(email),
			"profile":           graphql.String(profile),
			"role":              graphql.String("user"),
			"is_email_verified": graphql.Boolean(true),
			"googleID":          googleID,
			"githubID":          githubID,
		}

		err = client.Mutate(ctx, &mutation, mutationVars)
		if err != nil {
			log.Println("Something went wrong:", err.Error())
			return "", "", 0, "", err
		}

		// Fetch the newly created user
		err = client.Query(ctx, &query, queryVars)
		if err != nil {
			log.Println("Something went wrong:", err.Error())
			return "", "", 0, "", err
		}
	}

	user := query.User[0]

	// Generate tokens
	token, refreshToken, err = GenerateAllTokens(string(user.Email), string(user.UserName), string(user.Role), fmt.Sprintf("%d", user.ID), int(user.ID))
	if err != nil {
		log.Println("Something went wrong:", err.Error())
		return "", "", 0, "", err
	}

	return token, refreshToken, int(user.ID), string(user.Role), nil
}
