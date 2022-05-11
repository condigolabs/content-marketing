WITH D AS (
SELECT D.Uri , D.FullId
FROM `core-ssp.machinelearning.raw_articles_ext`  D
LEFT JOIN  `core-ssp.machinelearning.readable_article`   A ON A.Uri= D.Uri
WHERE A.Uri IS NULL )
SELECT Uri, ID as LabelId ,L.FullId , Label,Locale
FROM D
INNER JOIN `core-ssp.machinelearning.GPT_Labels` L ON L.FullId = D.FullId AND Locale= "en-US"