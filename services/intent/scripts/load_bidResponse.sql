
SELECT 
 RequestId
 ,MAX(Country)  as Country
 ,L.ID 
 ,MAX(Date)  AS DATE
 ,"{{.Locale}}" as Locale
 ,MAX(L.Label) as Label
 ,ARRAY_AGG( STRUCT (
 Products.Title
,Products.Description
,Products.Brand
,Products.Domain
,Products.Image
,Products.Target)) As Products
 FROM `core-ssp.machinelearning.usa_products_unnested` 
 INNER JOIN `core-ssp.machinelearning.GPT_Labels` L
 ON L.Label=  Products.Label
 AND Locale ="{{.Locale}}" AND L.ID >0
 WHERE RequestId ="{{.RequestId}}"
 GROUP BY RequestId, L.ID