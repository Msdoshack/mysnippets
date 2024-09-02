package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Snippet struct {
	ID      int
	Title   string
	Content string
	Language string
	Visibility string
	UserId int
	Name string
	Created time.Time
	Expires time.Time
}

type SnippetByLang map[string][]*Snippet

type SnippetModel struct {
	DB *sql.DB
}



func (m *SnippetModel) Insert(title, content, language, visibility string, userId int) (int, error) {
	stmt := `
		INSERT INTO snippets (title, content, language, visibility, "userId", created)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
		RETURNING id
	`

	// Use QueryRow to get the returning id
	var id int
	err := m.DB.QueryRow(stmt, title, content, language, visibility, userId).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}




func (m *SnippetModel) Get(id int) (*Snippet, error) {
	stmt := `
		SELECT 
			snippets.id, snippets.title, snippets.content, snippets.language, snippets.visibility, snippets."userId", 
			users.full_name AS name, snippets.created 
		FROM snippets 
		JOIN users ON snippets."userId" = users.id 
		WHERE snippets.id = $1
	`

	row := m.DB.QueryRow(stmt, id)

	s := &Snippet{}

	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Language, &s.Visibility, &s.UserId, &s.Name, &s.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrorNoRecord
		} else {
			return nil, err
		}
	}

	return s, nil
}

func (m *SnippetModel) GetLatest(userId, page int, title, language, sort string) ([]*Snippet, []string, error) {
	// Base query with a CTE (Common Table Expression) to get distinct languages and snippets in one go
	baseQuery := `
		WITH snippet_languages AS (
			SELECT DISTINCT language
			FROM snippets
			WHERE "userId" = $1
		)
		SELECT 
			snippets.id, snippets.title, snippets.content, snippets.language, snippets.visibility, snippets."userId", 
			users.full_name AS name, 
			snippets.created,
			(SELECT STRING_AGG(language, ',') FROM snippet_languages) AS languages
		FROM snippets
		JOIN users ON snippets."userId" = users.id
		WHERE users.id = $1
	`

	// Slice to store conditions dynamically
	conditions := []string{}

	params := []interface{}{userId,} // userId for both snippets and languages

	// Add language condition if provided
	if language != "" {
		conditions = append(conditions, "snippets.language = $"+fmt.Sprint(len(params)+1))
		params = append(params, language)
	}

	// Add title condition if provided
	if title != "" {
		conditions = append(conditions, "snippets.title ILIKE $"+fmt.Sprint(len(params)+1))
		// title = "%" + title + "%"
		params = append(params, "%"+title+"%")
	}

	// Combine base query with dynamic conditions
	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	offset := (page - 1) * 9

	if sort == "" {
		sort= "DESC"
	}

	// Add order by and limit clauses
	baseQuery += fmt.Sprintf(" ORDER BY snippets.created %s LIMIT 9 OFFSET %d", sort, offset)

	// Execute the query
	rows, err := m.DB.Query(baseQuery, params...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	// Snippets slice to store result
	snippets := []*Snippet{}
	var languages []string
	
	for rows.Next() {
		snippet := &Snippet{}
		var languageList string

		// Scan snippet data and comma-separated language list
		err := rows.Scan(&snippet.ID, &snippet.Title, &snippet.Content, &snippet.Language, &snippet.Visibility, &snippet.UserId, &snippet.Name, &snippet.Created, &languageList)
		if err != nil {
			return nil, nil, err
		}

		// Append snippet to slice
		snippets = append(snippets, snippet)

		// Split comma-separated languages and store them uniquely
		if len(languages) == 0 { // Process languages only once
			languages = strings.Split(languageList, ",")
		}
	}

	// Check for errors after looping through rows
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return snippets, languages, nil
}



func (m *SnippetModel) GetUserSnippets(id int) (SnippetByLang, error) {
	stmt := `
	  WITH RankedSnippets AS (
		SELECT 
			snippets.id, 
			title, 
			content, 
			language, 
			visibility, 
			"userId",
			users.full_name AS name, 
			snippets.created,
			ROW_NUMBER() OVER (PARTITION BY language ORDER BY snippets.created DESC) AS row_num
		FROM 
			snippets
		JOIN 
			users ON snippets."userId" = users.id 
		WHERE 
			"userId" = $1
	  )
	  SELECT 
		id, 
		title, 
		content, 
		language, 
		visibility, 
		"userId", 
		name,
		created
	  FROM 
		RankedSnippets
	  WHERE 
		row_num <= 6
	  ORDER BY 
		language ASC, 
		created DESC;
	`

	rows, err := m.DB.Query(stmt, id)
	if err != nil {
		return nil, fmt.Errorf("error executing query to get user snipper %w",err)
	}

	defer rows.Close()

	// A map to store snippets grouped by language
	snippetsByLanguage := make(map[string][]*Snippet)

	// Iterate over the rows
	for rows.Next() {
		var snippet Snippet

		if err := rows.Scan(&snippet.ID, &snippet.Title, &snippet.Content, &snippet.Language, &snippet.Visibility, &snippet.UserId, &snippet.Name, &snippet.Created); err != nil {
			return nil, err
		}

		// Group snippets by language
		snippetsByLanguage[snippet.Language] = append(snippetsByLanguage[snippet.Language], &snippet)
	}

	// Check for errors from row iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return snippetsByLanguage, nil
}



func (m *SnippetModel) UpdateSnippet(title string, content string, language string, visibility string, id int) error {
	stmt := `UPDATE snippets SET title = $1, content = $2, language = $3, visibility = $4 WHERE id = $5`

	_, err := m.DB.Exec(stmt, title, content, language, visibility, id)
	if err != nil {
		return err
	}

	return nil
}


func (m *SnippetModel) DelSnippet(id int) error {
	stmt := `DELETE FROM snippets WHERE id = $1`

	_, err := m.DB.Exec(stmt, id)
	if err != nil {
		return err
	}

	return nil
}