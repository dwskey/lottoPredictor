-- 회차별 번호 확률
SELECT * 
FROM draw_probabilities
WHERE draw_number = 1167
ORDER BY number 
;

-- 회차별 확률 높은 순서
SELECT * 
FROM draw_probabilities
WHERE draw_number = 1167
ORDER BY probability desc
;