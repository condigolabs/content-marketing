WITH D AS ( SELECT
    Date
    ,RequestId
    ,OrganizationId
    ,ApplicationId
    ,R.Domain
    ,Context
    ,Country
    ,ProductTokens[SAFE_OFFSET(0)] as Product
    ,(SELECT SAFE_DIVIDE(SUM(CPM),COUNT(*)) FROM UNNEST(Pricing) AS P) AS AvgBid
FROM `core-ssp.reporting.live_request_full`  as R
WHERE
    ResponseStatus IN(200)
AND
    date >  TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 120 SECOND)
AND
    DemandId ="criteortb"
)
SELECT * FROM D
WHERE Product is NOT NULL  AND Product.UniqueId NOT LIKE "fr-FR%"
ORDER BY Date DESC LIMIT 50