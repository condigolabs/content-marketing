package intent

import (
	"bytes"
	"cloud.google.com/go/bigquery"
	"context"
	"errors"
	"fmt"
	"github.com/condigolabs/content-marketing/models"
	"golang.org/x/net/html"
	"google.golang.org/api/iterator"
	"strings"
)

type LookupArticleParam struct {
	Fields     string
	QueryTerms string
	LabelId    int64
}

func (p *LookupArticleParam) CleanUp() {
	p.QueryTerms = strings.Replace(p.QueryTerms, "and", ",", -1)
	p.QueryTerms = strings.Replace(p.QueryTerms, "|", ",", -1)
	p.QueryTerms = strings.Replace(p.QueryTerms, ">", ",", -1)
	p.QueryTerms = strings.Replace(p.QueryTerms, "&", " ", -1)
}

func (dw *ConcreteIntent) LookupArticles(param LookupArticleParam) ([]models.Article, error) {
	var buf bytes.Buffer
	param.CleanUp()
	err := dw.templates.ExecuteTemplate(&buf, "lookup_articles.sql", param)

	if err != nil {
		return nil, err
	}
	fmt.Printf("***\n")
	fmt.Printf(buf.String())
	q := dw.client.Query(buf.String())
	q.Priority = bigquery.InteractivePriority
	q.QueryConfig.UseLegacySQL = false
	ctx := context.Background()
	// Start the job.
	job, err := q.Run(ctx)
	if err != nil {
		return nil, err
	}
	status, err := job.Wait(ctx)
	if err != nil {
		return nil, err
	}
	if err := status.Err(); err != nil {
		return nil, err
	}
	ret := make([]models.Article, 0)
	it, err := job.Read(ctx)
	for {
		var row models.Article
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			continue
		}
		// Deal With Content Manipulation
		s := strings.Split(row.TextContent, "\n")
		c := 0
		for _, l := range s {
			if len(l) > 0 {
				l = strings.Replace(l, "\t", "", -1)
				l1 := strings.Split(l, ".")
				for _, l := range l1 {
					if len(l) > 120 {
						row.Lines = append(row.Lines, fmt.Sprintf("%s", strings.Trim(l, " ")))
						c++
					}
					if c > 10 {
						break
					}
				}
			}
			if c > 10 {
				break
			}
		}
		//
		/*doc, err := html.Parse(strings.NewReader(row.TextContent))
		Div(doc)*/
		ret = append(ret, row)
	}
	return ret, err
}

func Div(doc *html.Node) (*html.Node, error) {
	var body *html.Node
	var crawler func(*html.Node)
	crawler = func(node *html.Node) {

		if node.Type == html.ElementNode && node.Data == "div" {
			body = node
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			crawler(child)
		}
	}
	crawler(doc)
	if body != nil {
		return body, nil
	}
	return nil, errors.New("Missing <body> in the node tree")
}

func (dw *ConcreteIntent) LookupArticlesNew(param LookupArticleParam) ([]models.ArticleData, error) {
	var buf bytes.Buffer
	param.CleanUp()
	err := dw.templates.ExecuteTemplate(&buf, "load_articles_new.sql", param)

	if err != nil {
		return nil, err
	}
	fmt.Printf("***\n")
	fmt.Printf(buf.String())
	q := dw.client.Query(buf.String())
	q.Priority = bigquery.InteractivePriority
	q.QueryConfig.UseLegacySQL = false
	ctx := context.Background()
	// Start the job.
	job, err := q.Run(ctx)
	if err != nil {
		return nil, err
	}
	status, err := job.Wait(ctx)
	if err != nil {
		return nil, err
	}
	if err := status.Err(); err != nil {
		return nil, err
	}
	ret := make([]models.ArticleData, 0)
	it, err := job.Read(ctx)
	for {
		var row models.ArticleData
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			continue
		}
		// Deal With Content Manipulation
		s := strings.Split(row.TextContent, "\n")
		c := 0
		for _, l := range s {
			if len(l) > 0 {
				l = strings.Replace(l, "\t", "", -1)
				l1 := strings.Split(l, ".")
				for _, l := range l1 {
					if len(l) > 50 {
						if c == 0 && row.Level == 1 {
							row.Title = fmt.Sprintf("%s", strings.Trim(l, " "))
						}
						row.Lines = append(row.Lines, fmt.Sprintf("%s", strings.Trim(l, " ")))
						c++
					}
					if c > 25 {
						break
					}
				}
			}
			if c > 25 {
				break
			}
		}
		ret = append(ret, row)
	}
	return ret, err
}
