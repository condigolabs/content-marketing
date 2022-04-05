WITH  I AS (
SELECT DISTINCT Label,LabelId, ImageUrls[SAFE_OFFSET(0)] as PublisherImage
FROM `core-ssp.adserver.creative_items_x`  ORDER BY RAND() LIMIT 1)
SELECT RequestId,InputText,Model,Language, Method, Description,Image, COALESCE(T.PublisherImage, I.PublisherImage) as PublisherImage FROM `core-ssp.machinelearning.generated_text` T, I
WHERE RequestId ="{{.}}"
