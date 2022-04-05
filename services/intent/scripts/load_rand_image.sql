SELECT DISTINCT Label,LabelId, ImageUrls[SAFE_OFFSET(0)] as Image
FROM `core-ssp.adserver.creative_items_x`
WHERE Label LIKE "%{{.Tag}}%"
ORDER BY RAND() LIMIT 5