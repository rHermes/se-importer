with months(n) as
         (
             select 1 as n
             union all
             select n + 1
             from months
             where n + 1 <= 12
         ),
     years(n) as
         (
             SELECT 2009 as n
             union all
             select n + 1
             from years
             where n + 1 <= 2019
         ),
     pp(year, month) as
         (
             select years.n  as [year],
                    months.n as [month]
             from years
                      CROSS JOIN months
         )
SELECT DATEFROMPARTS(pp.year, pp.month, 1)                                                           as day,
       count(u.id)                                                                                   as new,
       SUM(count(u.id)) OVER (ORDER BY DATEFROMPARTS(pp.year, pp.month, 1) ROWS UNBOUNDED PRECEDING) as cum
FROM pp
         LEFT OUTER JOIN [user] u ON (pp.year = YEAR(u.creation_date) AND pp.month = MONTH(u.creation_date))
GROUP BY DATEFROMPARTS(pp.year, pp.month, 1)
ORDER BY 1
option (maxrecursion 0);

with topexchanges as (
    SELECT TOP 10 s.name,
                  s.id
    FROM site s
             LEFT OUTER JOIN [user] u on s.id = u.site_id
    GROUP BY s.name, s.id
    ORDER BY count(u.id) DESC
),
     months(n) as
         (
             select 1 as n
             union all
             select n + 1
             from months
             where n + 1 <= 12
         ),
     years(n) as
         (
             SELECT 2009 as n
             union all
             select n + 1
             from years
             where n + 1 <= 2019
         ),
     pp(year, month) as
         (
             select years.n  as [year],
                    months.n as [month]
             from years
                      CROSS JOIN months
         )
SELECT s.name,
       DATEFROMPARTS(pp.year, pp.month, 1)                                                           as day,
       count(u.id)                                                                                   as new,
       SUM(count(u.id)) OVER (ORDER BY DATEFROMPARTS(pp.year, pp.month, 1) ROWS UNBOUNDED PRECEDING) as cum
FROM topexchanges s
         CROSS JOIN
     pp
         LEFT OUTER JOIN [user] u
                         ON (pp.year = YEAR(u.creation_date) AND pp.month = MONTH(u.creation_date) AND s.id = u.site_id)
GROUP BY s.name, DATEFROMPARTS(pp.year, pp.month, 1)
ORDER BY 1, 2
option (maxrecursion 0);