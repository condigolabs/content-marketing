SELECT
Country
,Locale
,COALESCE(L.Root,T.Label) as Root
,COALESCE(L.Main, L.Root,T.Label) as Cat
,COALESCE(L.Sub1,L.Main, L.Root)  as SubCat
,L.ID
,L.Label
,AVG(avgBid) as AvgBid
,SUM(Count) as Count
FROM `core-ssp.machinelearning.content_top`  T
INNER  JOIN  `core-ssp.machinelearning.GPT_Labels` L ON  T.ID =L.ID AND Locale="{{.Locale}}" AND Rank<3
WHERE Country ="{{.Country}}"
GROUP BY  Country,  Root,Cat,SubCat,ID,Locale,L.Label
HAVING COUNT(*) > 400
ORDER BY COUNT(*) DESC