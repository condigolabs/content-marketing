WITH R AS ( 
   SELECT
        DISTINCT RequestId ,L.ID
   FROM `core-ssp.machinelearning.usa_products_unnested`
   INNER JOIN `core-ssp.machinelearning.GPT_Labels` L ON L.Label=  Products.Label
   AND L.ID = {{.LabelId}} AND Locale ="{{.Locale}}"
   ORDER BY RAND() DESC  LIMIT 1
)
  SELECT 
  R.RequestId
  ,MAX(Country)  as Country
 ,L.ID  as LabelId
 ,MAX(Date)  AS DATE
 ,"{{.Locale}}" as Locale
 ,MAX(L.Label) as Label
 ,ARRAY_AGG( STRUCT (
 Products.Title
,Products.Description
,Products.Brand
,Products.Domain
,Products.Image
,Products.Target) ORDER BY RAND() LIMIT 2) As Products
 FROM `core-ssp.machinelearning.usa_products_unnested`  Q 
 INNER JOIN  R ON R.RequestId =  Q.RequestId
INNER JOIN `core-ssp.machinelearning.GPT_Labels` L ON L.ID=  R.ID
 GROUP BY RequestId, LabelId
