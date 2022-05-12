SELECT
"USA" as Country
,L.Locale
,COALESCE(L.Root,L.Label) as Root
,COALESCE(L.Main, L.Root,L.Label) as Cat
,COALESCE(L.Sub1,L.Main, L.Root)  as SubCat
,L.ID
,L.Label
,0.0 as AvgBid
,1000*COUNT(*) as Count
FROM `core-ssp.machinelearning.readable_article`  A
INNER  JOIN  `core-ssp.machinelearning.GPT_Labels` L ON  L.ID =A.LabelId	 AND L.Locale="en-US"
GROUP BY   Root,Cat,SubCat,ID,L.Locale,L.Label
ORDER BY Count DESC