WITH D AS ( SELECT
    date
    ,OrganizationId
    ,ApplicationId
    ,R.Domain
    ,Context
    ,UserValue
    ,UserType
    ,Country
    ,DemandId
    ,P.Domain as Adv
    ,P.UniqueId
    ,P.Title
    ,P.Label
    ,P.Brand
    ,P.Image
    ,P.Target
    ,P.Description
    ,(SELECT SAFE_DIVIDE(SUM(CPM),COUNT(*)) FROM UNNEST(Pricing) AS P) AS AvgBid
FROM `core-ssp.reporting.live_request_full`  as R,
UNNEST(ProductTokens) as P
WHERE  ResponseStatus IN(200)
AND date >  TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL {.Hours} HOUR)
AND DemandId ="criteortb"
AND P.Domain <> ""
/*AND OrganizationId ="c60ileukv5dosgvfgpi0" */
AND Label <> ""
)
, A AS (SELECT Country ,Adv, Title, Brand ,Image,Target, Label,AVg(AvgBid) as avgBid, COUNT(DISTINCT UserValue) as UniqueUserID , COUNT(*) as Count
FROM D   GROUP BY Country ,Adv, Title, Brand, Label,Image,Target )

SELECT A.*,L.ID, LFR.Label as LabelFullName, LFR.FriendlyName,RANK() OVER (PARTITION BY Brand ORDER BY RAND() ) AS ProductRank  FROM A
LEFT  JOIN  `core-ssp.machinelearning.GPT_Labels` L ON L.Label= A.Label
LEFT JOIN  `core-ssp.machinelearning.GPT_Label_Matcher` M ON M.Label = A.InitialLabel
LEFT JOIN  `core-ssp.machinelearning.GPT_Labels`  LFR ON LFR.ID = L.ID AND LFR.Locale= "{.Locale}"