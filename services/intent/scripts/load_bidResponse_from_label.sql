WITH R AS ( 
   SELECT
        R.*,L.ID as LabelId, L.Label
   FROM `core-ssp.machinelearning.usa_products_unnested` R
   INNER JOIN `core-ssp.machinelearning.GPT_Labels` L ON L.Label=  Products.Label
   AND L.ID = {{.LabelId}} AND Locale ="{{.Locale}}"
   ORDER BY RAND() DESC  LIMIT 3
)
  SELECT
   MAX(RequestId) as RequestId
  ,MAX(Country)  as Country
  ,LabelId
  ,MAX(Date)  AS DATE
  ,"en-US" as Locale
  ,MAX(Label) as Label
  ,ARRAY_AGG( STRUCT (
  Products.Label
  ,Products.Title
  ,Products.Description
  ,Products.Brand
  ,Products.Domain
  ,Products.Image
  ,Products.Target) ) As Products
 FROM R

 GROUP BY  LabelId