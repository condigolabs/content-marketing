WITH A AS (
  SELECT
   Subject as Title
  ,GeneratedText as TextContent
  ,"" as Excerpt
  ,LabelId
  ,Label
  ,1 as Level
  FROM `core-ssp.machinelearning.generated_articles` A
  WHERE LabelId= {{.LabelId}}
  UNION ALL
  SELECT
    Title
    ,TextContent
    ,Excerpt
    ,LabelId
    ,Label
    ,2 as Level
    FROM `core-ssp.machinelearning.readable_article`
  WHERE LabelId= {{.LabelId}} OR SEARCH((Title,Label,TextContent ), '{{.QueryTerms}}')  AND Image <> ""
)
SELECT Title,TextContent,A.LabelId,A.Label, Level,
ARRAY_AGG(I.Urls ORDER BY RAND() DESC LIMIT 1)[OFFSET(0)] as Urls
FROM   A
INNER JOIN `core-ssp.machinelearning.unsplash_images`  I ON A.LabelId = I.LabelId
GROUP BY Title,TextContent,A.LabelId,A.Label, Level
ORDER BY level
LIMIT 2