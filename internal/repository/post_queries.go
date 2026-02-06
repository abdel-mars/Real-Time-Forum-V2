package repository

const (
	// INSERT queries
	INSERT_NEW_POST = `
                            INSERT INTO posts (user_id, title, content) VALUES (?, ?, ?);
                            UPDATE post_metadata SET post_count = post_count + 1;`
	MAP_POSTS_WITH_CATEGORY = `INSERT INTO post_categories (post_id, category_id) VALUES (?, ?);
                            UPDATE categories_count SET post_count = post_count + 1 WHERE category_id = ?;`

	INSERT_NEW_COMMENT         = `INSERT INTO comments (user_id, post_id, comment) VALUES (?, ?, ?)`
	INIT_POST_META_DATA        = `INSERT OR IGNORE INTO post_metadata (id, post_count) VALUES (1, 0);`
	INIT_FIELDS_QUERY          = `INSERT OR IGNORE INTO categories (name) VALUES (?)`
	INIT_POST_CATEGORIES_COUNT = `INSERT OR IGNORE INTO categories_count (category_id, post_count) VALUES (?, 0)`
	INSERT_NEW_LIKE_DISLIKE    = `INSERT INTO likes_dislikes (user_id, post_id, is_like, is_dislike) VALUES (?, ?, ?, ?)`

	// DELETE queries
	// todo: update the post categories count on delete user
	DELETE_POST = `
                DELETE FROM posts WHERE id = ?;
                UPDATE post_metadata SET post_count = post_count - 1;`

	// SELECT queries
	GET_POST_COUNT_BY_CAT  = `SELECT pcc.post_count FROM categories_count pcc JOIN categories c ON pcc.category_id = c.id WHERE c.name = ?`
	IS_POST_EXIST          = `SELECT 1 FROM posts WHERE id = ? LIMIT 1`
	IS_LIKED               = `SELECT is_like FROM likes_dislikes WHERE user_id = ? AND post_id = ?`
	IS_DISLIKED            = `SELECT is_dislike FROM likes_dislikes WHERE user_id = ? AND post_id = ?`
	SELECT_TODAY_POSTS     = `SELECT COUNT(*) FROM posts WHERE user_id = ?  AND created_at >= DATE('now')`
	SELECT_CATEGORY_ID     = `SELECT id FROM categories WHERE name = ?`
	GET_POST_COUNT         = `SELECT post_count FROM post_metadata`
	GET_OWNED_POST_COUNT   = `SELECT COUNT(*) FROM posts WHERE user_id = ?`
	GET_LIKED_POST_COUNT   = `SELECT COUNT(*) FROM likes_dislikes WHERE user_id = ? AND is_like == 1`
	GET_COMMENT_POST_COUNT = `SELECT COUNT(*) FROM comments WHERE post_id == ?`
	SELECT_TODAY_COMMENTS  = `SELECT COUNT(*) FROM comments WHERE user_id = ?  AND created_at >= DATE('now')`

	// UDDATE queries
	UPDATE_LIKE    = `UPDATE likes_dislikes SET is_like = ?, is_dislike = 0 WHERE user_id = ? AND post_id = ?`
	UPDATE_DISLIKE = `UPDATE likes_dislikes SET is_like = 0, is_dislike = ? WHERE user_id = ? AND post_id = ?`

	GET_POST_BYLIKES = `
  SELECT
    p.id,
    p.user_id,
    p.title,
    p.content,
    u.username AS publisher,
    IFNULL(GROUP_CONCAT(DISTINCT c.name), '') AS categories,
    COUNT(DISTINCT CASE WHEN ld.is_like = 1 THEN ld.user_id END) AS likes,
    COUNT(DISTINCT CASE WHEN ld.is_dislike = 1 THEN ld.user_id END) AS dislikes,
    p.created_at,
    p.updated_at
    FROM posts p
    JOIN users u ON u.id = p.user_id
    LEFT JOIN post_categories pc ON pc.post_id = p.id
    LEFT JOIN categories c ON c.id = pc.category_id
    LEFT JOIN likes_dislikes ld ON ld.post_id = p.id
    WHERE p.id IN (
    SELECT post_id
    FROM likes_dislikes
    WHERE user_id = ? AND is_like = 1
    )
    GROUP BY p.id
    ORDER BY p.created_at DESC
    LIMIT ? OFFSET ?;
  `
	GET_POST_BYOWNED = `
  SELECT 
    p.id,
    p.user_id,
    p.title,
    p.content,
    u.username AS publisher,
    IFNULL(GROUP_CONCAT(DISTINCT c.name), '') AS categories,
    COUNT(DISTINCT CASE WHEN ld.is_like = 1 THEN ld.user_id END) AS likes,
    COUNT(DISTINCT CASE WHEN ld.is_dislike = 1 THEN ld.user_id END) AS dislikes,
    p.created_at,
    p.updated_at
    FROM posts p
    JOIN users u ON u.id = p.user_id
    LEFT JOIN post_categories pc ON pc.post_id = p.id
    LEFT JOIN categories c ON c.id = pc.category_id
    LEFT JOIN likes_dislikes ld ON ld.post_id = p.id
    WHERE p.user_id = ?
    GROUP BY p.id
    ORDER BY p.created_at DESC
    LIMIT ? OFFSET ?;
  `
	GET_POST_BYCATEGORY = `
  SELECT
    p.id,
    p.user_id,
    p.title,
    p.content,
    u.username AS publisher,
    IFNULL(GROUP_CONCAT(DISTINCT c.name), '') AS categories,
    COUNT(DISTINCT CASE WHEN ld.is_like = 1 THEN ld.user_id END) AS likes,
    COUNT(DISTINCT CASE WHEN ld.is_dislike = 1 THEN ld.user_id END) AS dislikes,
    p.created_at,
    p.updated_at
    FROM posts p
    JOIN users u ON u.id = p.user_id
    LEFT JOIN post_categories pc ON pc.post_id = p.id
    LEFT JOIN categories c ON c.id = pc.category_id
    LEFT JOIN likes_dislikes ld ON ld.post_id = p.id
    WHERE c.name = ?
    GROUP BY p.id
    ORDER BY p.created_at DESC
    LIMIT ? OFFSET ?;
  `

	SELECT_ALL_POSTS = `
  SELECT 
    p.id,
    p.user_id,
    p.title,
    p.content,
    u.username AS publisher,
    IFNULL(GROUP_CONCAT(DISTINCT c.name), '') AS categories,
    COUNT(DISTINCT CASE WHEN ld.is_like = 1 THEN ld.user_id END) AS likes,
    COUNT(DISTINCT CASE WHEN ld.is_dislike = 1 THEN ld.user_id END) AS dislikes,
    p.created_at,
    p.updated_at
    FROM posts p
    JOIN users u ON u.id = p.user_id
    LEFT JOIN post_categories pc ON pc.post_id = p.id
    LEFT JOIN categories c ON c.id = pc.category_id
    LEFT JOIN likes_dislikes ld ON ld.post_id = p.id
    GROUP BY p.id
    ORDER BY p.created_at DESC
    LIMIT ? OFFSET ?;
  `

	SELECT_POST_BY_ID = `
  SELECT 
    p.id, p.user_id, p.title, p.content, u.username AS publisher,
    IFNULL(GROUP_CONCAT(DISTINCT c.name), '') AS categories,
    COUNT(DISTINCT CASE WHEN ld.is_like = 1 THEN ld.user_id END) AS likes,
    COUNT(DISTINCT CASE WHEN ld.is_dislike = 1 THEN ld.user_id END) AS dislikes,
    p.created_at, p.updated_at
    FROM posts p
    JOIN users u ON u.id = p.user_id
    LEFT JOIN post_categories pc ON pc.post_id = p.id
    LEFT JOIN categories c ON c.id = pc.category_id
    LEFT JOIN likes_dislikes ld ON ld.post_id = p.id
    WHERE p.id = ?
    GROUP BY p.id;
  `
	SELECT_COMMENT_BY_10 = `
  SELECT users.username, comments.comment, comments.id
    FROM comments
    JOIN users ON comments.user_id = users.id
    WHERE comments.post_id = ?
    ORDER BY comments.created_at DESC
    LIMIT ? OFFSET ?;
  `
	IS_COMMENT_EXIST = `SELECT 1 FROM comments WHERE post_id = ? AND id = ?`


	IS_COMMENT_LIKED = `SELECT is_like FROM comment_likes_dislikes WHERE user_id = ? AND comment_id = ?;`
	IS_COMMENT_DISLIKED = `SELECT is_dislike FROM comment_likes_dislikes WHERE user_id = ? AND comment_id = ?;`




	INSERT_COMMENT_LIKE_DISLIKE = `
  INSERT INTO comment_likes_dislikes
      (user_id, comment_id, is_like, is_dislike, created_at)
  VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP);`



	UPDATE_COMMENT_LIKE = `
  UPDATE comment_likes_dislikes
     SET is_like    = ?,
         is_dislike = 0
   WHERE user_id    = ?
     AND comment_id = ?;`

	UPDATE_COMMENT_DISLIKE = `
  UPDATE comment_likes_dislikes
     SET is_like    = 0,
         is_dislike = ?
   WHERE user_id    = ?
     AND comment_id = ?;`

  GET_COMMENT_LIKE_COUNT = `SELECT COUNT(*) FROM comment_likes_dislikes WHERE comment_id = ? AND is_like = 1`
  GET_COMMENT_DISLIKE_COUNT =  `SELECT COUNT(*) FROM comment_likes_dislikes WHERE comment_id = ? AND is_dislike = 1`

  GET_CAT_COUNT = `
  SELECT COUNT(*) AS post_count
  FROM post_categories pc
  JOIN categories c ON pc.category_id = c.id
  WHERE c.name = ?;
  `
  UPDATE_CAT_COUNT = `UPDATE categories_count SET post_count = ? WHERE category_id = ?; `

)
