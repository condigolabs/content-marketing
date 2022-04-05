SELECT
Country
,COALESCE(L.Root,T.InitialLabel) as Cat
/*,COALESCE(L.Main, L.Root,T.InitialLabel) as Cat*/
/*,COALESCE(L.Sub1,L.Main, L.Root)  as SubCat*/
,AVG(avgBid) as AvgBid
,SUM(Count) as Count
FROM `core-ssp.machinelearning.content_top`  T
INNER  JOIN  `core-ssp.machinelearning.GPT_Labels` L ON  T.ID =L.ID AND Locale="{{.Locale}}"
GROUP BY  Country,  Cat
HAVING COUNT(*) > 400
ORDER BY COUNT(*) DESC