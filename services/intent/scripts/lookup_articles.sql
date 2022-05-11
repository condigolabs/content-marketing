SELECT Domain, Title,TextContent,Content,Excerpt,Image,Label FROM `core-ssp.machinelearning.readable_article`
WHERE SEARCH(({{.Fields}}), '{{.QueryTerms}}')
LIMIT 4