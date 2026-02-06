package db

import (
	"forum/internal/utils"
	"database/sql"
	"strings"
	repo "forum/internal/repository"
)

func AddNewPost(userId int, titel string, content string) (int, error) {
	res, err := repo.DB.Exec(repo.INSERT_NEW_POST, userId, titel, content)
	if err != nil {
		return -1, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func GetAllPostsInfo(page int, userId int) (repo.PageData, error) {
	var data repo.PageData
	rows, err := repo.DB.Query(repo.SELECT_ALL_POSTS, repo.PAGE_POSTS_QUANTITY, (page-1)*repo.PAGE_POSTS_QUANTITY)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		var post repo.Post
		var categoriesStr string

		err := rows.Scan(
			&post.Id,
			&post.PublisherId,
			&post.Title,
			&post.Content,
			&post.Publisher,
			&categoriesStr,
			&post.Likes,
			&post.Dislikes,
			&post.Created_at,
			&post.Updated_at,
		)
		if err != nil {
			return data, err
		}
		if categoriesStr != "" {
			post.Catigories = strings.Split(categoriesStr, ",")
		}

		post.IsEdited = post.Created_at != post.Updated_at
		post.IsLikedByUser, err = IsPostLikedByUser(userId, post.Id)
		if err != nil && err != sql.ErrNoRows {
			return data, err
		}
		post.IsDislikedByUser, err = IsPostDisikedByUser(userId, post.Id)
		if err != nil && err != sql.ErrNoRows {
			return data, err
		}
		if len(post.Catigories) != 0 {
			post.HasCategories = true
		}
		if userId == post.PublisherId {
			post.Owned = true
		}
		post.CommentsCount, err = GetCommentCount(post.Id)
		if err != nil {
			return data, err
		}
		if post.Publisher != "" {
			post.Initial = post.Publisher[:1]
		}
		if len(post.Catigories) != 0 {
			post.HasCategories = true
		}
		post.Created_at = utils.SqlDateFormater(post.Created_at)
		post.Updated_at = utils.SqlDateFormater(post.Updated_at)
		data.Posts = append(data.Posts, post)
	}

	if err := rows.Err(); err != nil {
		return data, err
	}

	return data, nil
}

func IsPostExist(postId int) (bool, error) {
	var exists bool

	err := repo.DB.QueryRow(repo.IS_POST_EXIST, postId).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func AddRemovePostLike(userId int, postId int) error {
	isLiked, err := IsPostLikedByUser(userId, postId)
	if err == sql.ErrNoRows {
		_, err = repo.DB.Exec(repo.INSERT_NEW_LIKE_DISLIKE, userId, postId, 1, 0)
		return err
	}
	if isLiked {
		res, err := repo.DB.Exec(repo.UPDATE_LIKE, 0, userId, postId)
		if err != nil {
			return err
		}
		_, err = res.RowsAffected()
		return err
	}
	res, err := repo.DB.Exec(repo.UPDATE_LIKE, 1, userId, postId)
	if err != nil {
		return err
	}
	_, err = res.RowsAffected()
	return err
}

func AddRemovePostDeslike(userId int, postId int) error {
	isDisike, err := IsPostDisikedByUser(userId, postId)
	if err == sql.ErrNoRows {
		_, err = repo.DB.Exec(repo.INSERT_NEW_LIKE_DISLIKE, userId, postId, 0, 1)
		return err
	}
	if isDisike {
		res, err := repo.DB.Exec(repo.UPDATE_DISLIKE, 0, userId, postId)
		if err != nil {
			return err
		}
		_, err = res.RowsAffected()
		return err
	}
	res, err := repo.DB.Exec(repo.UPDATE_DISLIKE, 1, userId, postId)
	if err != nil {
		return err
	}
	_, err = res.RowsAffected()
	return err
}

func IsPostLikedByUser(userId int, postId int) (bool, error) {
	var isLike bool

	err := repo.DB.QueryRow(repo.IS_LIKED, userId, postId).Scan(&isLike)
	if err == sql.ErrNoRows {
		return false, err
	} else if err != nil {
		return false, err
	} else if isLike {
		return true, nil
	}
	return false, nil
}

func IsPostDisikedByUser(userId int, postId int) (bool, error) {
	var isDisike bool

	err := repo.DB.QueryRow(repo.IS_DISLIKED, userId, postId).Scan(&isDisike)
	if err == sql.ErrNoRows {
		return false, err
	} else if err != nil {
		return false, err
	} else if isDisike {
		return true, nil
	}
	return false, nil
}

func MapPostWithCategories(postId int, categories []string) error {
	for _, category := range categories {
		categoryId, err := GetCategoryId(category)
		if err != nil {
			return err
		}
		MapPostCategory(postId, categoryId)
	}
	return nil
}

func GetCategoryId(category string) (int, error) {
	var categoryId int

	err := repo.DB.QueryRow(repo.SELECT_CATEGORY_ID, category).Scan(&categoryId)
	if err == sql.ErrNoRows {
		return -1, err
	} else if err != nil {
		return -1, err
	}
	return categoryId, nil
}

func MapPostCategory(postId int, categoryId int) error {
	_, err := repo.DB.Exec(repo.MAP_POSTS_WITH_CATEGORY, postId, categoryId, categoryId)
	return err
}

func GetPostsCount(filter string, userId int) (int, error) {
	var count int
	var query string
	var param any

	if filter == "" {
		err := repo.DB.QueryRow(repo.GET_POST_COUNT).Scan(&count)
		if err == sql.ErrNoRows {
			return 0, nil
		} else if err != nil {
			return -1, err
		}
		return count, nil
	}

	if filter == "Owned" {
		query = repo.GET_OWNED_POST_COUNT
		param = userId

	} else if filter == "Likes" {
		query = repo.GET_LIKED_POST_COUNT
		param = userId

	} else if repo.IT_MAJOR_FIELDS[filter] {
		query = repo.GET_POST_COUNT_BY_CAT
		param = filter
	}

	err := repo.DB.QueryRow(query, param).Scan(&count)

	if err == sql.ErrNoRows {
		return 0, nil
	} else if err != nil {
		return -1, err
	}
	return count, nil

}

func IsUserCanPostToday(userId int) (bool, error) {
	var postCount int

	err := repo.DB.QueryRow(repo.SELECT_TODAY_POSTS, userId).Scan(&postCount)
	if err != nil {
		return false, err
	}
	return postCount < repo.DAY_POST_LIMIT, nil
}

func GetPostByID(postID, userID int) (repo.Post, error) {
	var post repo.Post
	var categoriesStr string

	row := repo.DB.QueryRow(repo.SELECT_POST_BY_ID, postID)

	err := row.Scan(
		&post.Id,
		&post.PublisherId,
		&post.Title,
		&post.Content,
		&post.Publisher,
		&categoriesStr,
		&post.Likes,
		&post.Dislikes,
		&post.Created_at,
		&post.Updated_at,
	)
	if err != nil {
		return post, err
	}

	if categoriesStr != "" {
		post.Catigories = strings.Split(categoriesStr, ",")
	}

	post.IsEdited = post.Created_at != post.Updated_at
	post.IsLikedByUser, err = IsPostLikedByUser(userID, post.Id)
	if err != nil && err != sql.ErrNoRows {
		return post, err
	}
	post.IsDislikedByUser, err = IsPostDisikedByUser(userID, post.Id)
	if err != nil && err != sql.ErrNoRows {
		return post, err
	}
	if post.Publisher != "" {
		post.Initial = post.Publisher[:1]
	}
	if len(post.Catigories) != 0 {
		post.HasCategories = true
	}
	post.Created_at = utils.SqlDateFormater(post.Created_at)
	post.Updated_at = utils.SqlDateFormater(post.Updated_at)
	return post, nil
}

func UpdateCatCount() error {
	for k, _ := range repo.IT_MAJOR_FIELDS {
		if k != "All categories" {
			catId, err := GetCategoryId(k)
			if err != nil {
				return err
			}
			count, err := GetCatCount(catId)
			if err != nil {
				return err
			}
			_, err = repo.DB.Exec(repo.UPDATE_CAT_COUNT, count, catId)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func GetCatCount(catId int) (int, error) {
	var count int

	err := repo.DB.QueryRow(repo.GET_CAT_COUNT, catId).Scan(&count)

	if err == sql.ErrNoRows {
		return 0, nil
	} else if err != nil {
		return -1, err
	}
	return count, nil

}
