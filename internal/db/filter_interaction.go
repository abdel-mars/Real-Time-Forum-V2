package db

import (
	repo "forum/internal/repository"
	"database/sql"
	"strings"
	"forum/internal/utils"
)

func Getposbytlikes(userId int, page int) (repo.PageData, error) {
	var data repo.PageData

	rows, err := repo.DB.Query(repo.GET_POST_BYLIKES, userId, repo.PAGE_POSTS_QUANTITY, (page-1)*repo.PAGE_POSTS_QUANTITY)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		var post repo.Post
		var categoriesStr string

		err := rows.Scan(
			&post.Id,
			&post.Publisher,
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
		post.IsLikedByUser = true
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

func Getpostbyowner(userId int, page int) (repo.PageData, error) {
	var data repo.PageData

	rows, err := repo.DB.Query(repo.GET_POST_BYOWNED, userId, repo.PAGE_POSTS_QUANTITY, (page-1)*repo.PAGE_POSTS_QUANTITY)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		var post repo.Post
		var categoriesStr string

		err := rows.Scan(
			&post.Id,
			&post.Publisher,
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
		post.Created_at = utils.SqlDateFormater(post.Created_at)
		post.Updated_at = utils.SqlDateFormater(post.Updated_at)
		data.Posts = append(data.Posts, post)
	}

	if err := rows.Err(); err != nil {
		return data, err
	}
	return data, nil
}

func GePostbycategory(category string, page int, userId int) (repo.PageData, error) {
	var data repo.PageData

	rows, err := repo.DB.Query(repo.GET_POST_BYCATEGORY, category, repo.PAGE_POSTS_QUANTITY, (page-1)*repo.PAGE_POSTS_QUANTITY)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		var post repo.Post
		var categoriesStr string

		err := rows.Scan(
			&post.Id,
			&post.Publisher,
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
