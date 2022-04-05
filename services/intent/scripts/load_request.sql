WITH A AS (
SELECT
R.Domain as PublisherDomain
,Context as Page
,Label as InitialLabel
,Product.Title
,Product.Description
,Product.Brand
,Product.Domain
,Product.Image
,Product.Target

FROM `core-ssp.reporting.live_request_full`  R , UNNEST(ProductTokens) as Product
WHERE Date(Date) = CURRENT_DATE() AND RequestId ="{{.RequestId}}")
, U  AS (SELECT A.*, COALESCE(LFR.Label, A.InitialLabel)  as Label FROM A
LEFT  JOIN  `core-ssp.machinelearning.GPT_Labels` L ON L.Label= A.InitialLabel
LEFT JOIN  `core-ssp.machinelearning.GPT_Label_Matcher` M ON M.Label = A.InitialLabel
LEFT JOIN  `core-ssp.machinelearning.GPT_Labels`  LFR ON LFR.ID = COALESCE(L.ID, M.ID) AND LFR.Locale= "{{.Locale}}")
SELECT DISTINCT Title, Brand,Domain,Image,Target FROM U
LIMIT 5