WITH D AS ( SELECT
    R.Domain
    ,COUNT(*) as Count
FROM `core-ssp.reporting.live_request_full`  as R
WHERE
ResponseStatus IN(200)
AND date >  TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 120 SECOND)
AND DemandId ="criteortb"
GROUP BY R.Domain
)
SELECT * FROM D ORDER BY Count DESC