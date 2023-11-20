// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// move to API test
// func TestCreateBookmarkEvent(t *testing.T) {
// 	th := Setup(t).InitBasic()
// 	defer th.TearDown()

// 	WebSocketClient, err := th.CreateWebSocketClient()
// 	require.NoError(t, err)
// 	WebSocketClient.Listen()
// 	defer WebSocketClient.Close()

// 	bookmark1 := &model.ChannelBookmark{
// 		ChannelId:   th.BasicChannel.Id,
// 		DisplayName: "Link bookmark test",
// 		LinkUrl:     "https://mattermost.com",
// 		Type:        model.ChannelBookmarkLink,
// 		Emoji:       ":smile:",
// 	}

// 	th.Context.Session().UserId = th.BasicUser.Id // set the user for the session
// 	_, err = th.App.CreateChannelBookmark(th.Context, bookmark1, "")
// 	require.NoError(t, err)

// 	var received bool
// 	var b model.ChannelBookmarkWithFileInfo

// 	for {
// 		var exit bool
// 		select {
// 		case event := <-WebSocketClient.EventChannel:
// 			if event.EventType() == model.WebsocketEventChannelBookmarkCreated {
// 				err := json.Unmarshal([]byte(event.GetData()["bookmakr"].(string)), &b)
// 				require.NoError(t, err)
// 				received = true
// 			}
// 		case <-time.After(2 * time.Second):
// 			exit = true
// 		}
// 		if exit {
// 			break
// 		}
// 	}

// 	require.True(t, received)
// 	assert.NotNil(t, b)
// 	assert.NotEmpty(t, b.Id)
// }

func find_bookmark(slice []*model.ChannelBookmarkWithFileInfo, id string) *model.ChannelBookmarkWithFileInfo {
	for _, element := range slice {
		if element.Id == id {
			return element
		}
	}
	return nil
}

func TestCreateBookmark(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	t.Run("create a channel bookmark", func(t *testing.T) {
		bookmark1 := &model.ChannelBookmark{
			ChannelId:   th.BasicChannel.Id,
			DisplayName: "Link bookmark test",
			LinkUrl:     "https://mattermost.com",
			Type:        model.ChannelBookmarkLink,
			Emoji:       ":smile:",
		}

		th.Context.Session().UserId = th.BasicUser.Id // set the user for the session
		bookmarkResp, err := th.App.CreateChannelBookmark(th.Context, bookmark1, "")
		assert.Nil(t, err)

		assert.Equal(t, bookmarkResp.ChannelId, th.BasicChannel.Id)
		assert.NotEmpty(t, bookmarkResp.Id)

		bookmark2 := &model.ChannelBookmark{
			ChannelId:   th.BasicChannel.Id,
			OwnerId:     th.BasicUser.Id,
			DisplayName: "Link bookmark test",
			LinkUrl:     "https://mattermost.com",
			Type:        model.ChannelBookmarkFile,
			Emoji:       ":smile:",
		}
		bookmarkResp, err = th.App.CreateChannelBookmark(th.Context, bookmark2, "")
		assert.Nil(t, bookmarkResp)
		assert.NotNil(t, err)
	})
}

func TestUpdateBookmark(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	var updateBookmark *model.ChannelBookmark

	t.Run("same user update a channel bookmark", func(t *testing.T) {
		bookmark1 := &model.ChannelBookmark{
			ChannelId:   th.BasicChannel.Id,
			DisplayName: "Link bookmark test",
			LinkUrl:     "https://mattermost.com",
			Type:        model.ChannelBookmarkLink,
			Emoji:       ":smile:",
		}

		th.Context.Session().UserId = th.BasicUser.Id // set the user for the session
		bookmarkResp, err := th.App.CreateChannelBookmark(th.Context, bookmark1, "")
		assert.Nil(t, err)

		updateBookmark = bookmarkResp.Clone()
		updateBookmark.DisplayName = "New name"
		response, _ := th.App.UpdateChannelBookmark(th.Context, updateBookmark, "")
		assert.Greater(t, response.Updated.UpdateAt, response.Updated.CreateAt)
	})

	t.Run("another user update a channel bookmark", func(t *testing.T) {
		updateBookmark2 := updateBookmark.Clone()
		th.Context.Session().UserId = th.BasicUser2.Id
		response, _ := th.App.UpdateChannelBookmark(th.Context, updateBookmark2, "")
		assert.Equal(t, response.Updated.OriginalId, response.Deleted.Id)
		assert.Equal(t, response.Updated.DeleteAt, int64(0))
		assert.Greater(t, response.Deleted.DeleteAt, int64(0))
	})

	t.Run("update an already deleted channel bookmark", func(t *testing.T) {
		bookmark1 := &model.ChannelBookmark{
			ChannelId:   th.BasicChannel.Id,
			DisplayName: "Link bookmark test",
			LinkUrl:     "https://mattermost.com",
			Type:        model.ChannelBookmarkLink,
			Emoji:       ":smile:",
		}

		th.Context.Session().UserId = th.BasicUser.Id // set the user for the session
		bookmarkResp, err := th.App.CreateChannelBookmark(th.Context, bookmark1, "")
		assert.Nil(t, err)

		updateBookmark = bookmarkResp.Clone()
		_, err = th.App.DeleteChannelBookmark(updateBookmark.Id, "")
		assert.Nil(t, err)

		updateBookmark.DisplayName = "New name"
		_, err = th.App.UpdateChannelBookmark(th.Context, updateBookmark, "")
		assert.NotNil(t, err)
	})
}

func TestDeleteBookmark(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	t.Run("delete a channel bookmark", func(t *testing.T) {
		bookmark1 := &model.ChannelBookmark{
			ChannelId:   th.BasicChannel.Id,
			DisplayName: "Link bookmark test",
			LinkUrl:     "https://mattermost.com",
			Type:        model.ChannelBookmarkLink,
			Emoji:       ":smile:",
		}

		th.Context.Session().UserId = th.BasicUser.Id // set the user for the session
		bookmarkResp, err := th.App.CreateChannelBookmark(th.Context, bookmark1, "")
		assert.Nil(t, err)

		bookmarkResp, err = th.App.DeleteChannelBookmark(bookmarkResp.Id, "")
		assert.Nil(t, err)
		assert.Greater(t, bookmarkResp.DeleteAt, int64(0))
	})
}

func TestGetAllChannelsBookmarks(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	th.Context.Session().UserId = th.BasicUser.Id // set the user for the session

	bookmark1 := &model.ChannelBookmark{
		ChannelId:   th.BasicChannel.Id,
		DisplayName: "Bookmark 1",
		LinkUrl:     "https://mattermost.com",
		Type:        model.ChannelBookmarkLink,
		Emoji:       ":smile:",
	}

	th.App.CreateChannelBookmark(th.Context, bookmark1, "")

	file := &model.FileInfo{
		Id:              model.NewId(),
		CreatorId:       th.BasicUser.Id,
		Path:            "somepath",
		ThumbnailPath:   "thumbpath",
		PreviewPath:     "prevPath",
		Name:            "test file",
		Extension:       "png",
		MimeType:        "images/png",
		Size:            873182,
		Width:           3076,
		Height:          2200,
		HasPreviewImage: true,
	}

	th.App.Srv().Store().FileInfo().Save(file)
	defer th.App.Srv().Store().FileInfo().PermanentDelete(file.Id)

	bookmark2 := &model.ChannelBookmark{
		ChannelId:   th.BasicChannel.Id,
		DisplayName: "Bookmark 2",
		FileId:      file.Id,
		Type:        model.ChannelBookmarkFile,
		Emoji:       ":smile:",
	}

	th.App.CreateChannelBookmark(th.Context, bookmark2, "")

	channel2 := th.CreateChannel(th.Context, th.BasicTeam)
	bookmark3 := &model.ChannelBookmark{
		ChannelId:   channel2.Id,
		DisplayName: "Bookmark 3",
		LinkUrl:     "https://mattermost.com",
		Type:        model.ChannelBookmarkLink,
	}

	th.App.CreateChannelBookmark(th.Context, bookmark3, "")

	bookmark4 := &model.ChannelBookmark{
		ChannelId:   channel2.Id,
		DisplayName: "Bookmark 4",
		LinkUrl:     "https://mattermost.com",
		Type:        model.ChannelBookmarkLink,
	}

	th.App.CreateChannelBookmark(th.Context, bookmark4, "")

	t.Run("get bookmarks on all channels", func(t *testing.T) {
		channelIds := []string{th.BasicChannel.Id, channel2.Id}
		bookmarks, err := th.App.GetAllChannelBookmarks(channelIds, 0)
		assert.Nil(t, err)
		assert.Len(t, bookmarks, 2)
		assert.Len(t, bookmarks[th.BasicChannel.Id], 2)
		assert.Len(t, bookmarks[channel2.Id], 2)
	})

	t.Run("get bookmarks on all channels after one is deleted (aka only return the changed bookmarks)", func(t *testing.T) {
		now := model.GetMillis()
		channelIds := []string{th.BasicChannel.Id, channel2.Id}

		th.App.DeleteChannelBookmark(bookmark3.Id, "")

		bookmarks, err := th.App.GetAllChannelBookmarks(channelIds, 0)
		assert.Nil(t, err)
		assert.Len(t, bookmarks, 2)
		assert.Len(t, bookmarks[th.BasicChannel.Id], 2)
		assert.Len(t, bookmarks[channel2.Id], 1)

		bookmarks, err = th.App.GetAllChannelBookmarks(channelIds, now)
		assert.Nil(t, err)
		assert.Len(t, bookmarks, 1)
		assert.Len(t, bookmarks[th.BasicChannel.Id], 0)
		assert.Len(t, bookmarks[channel2.Id], 1)

		deleted := false
		for _, b := range bookmarks[channel2.Id] {
			if b.DeleteAt > 0 {
				deleted = true
				break
			}
		}
		assert.Equal(t, deleted, true)
	})
}

func TestGetChannelBookmarks(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	th.Context.Session().UserId = th.BasicUser.Id // set the user for the session

	bookmark1 := &model.ChannelBookmark{
		ChannelId:   th.BasicChannel.Id,
		DisplayName: "Bookmark 1",
		LinkUrl:     "https://mattermost.com",
		Type:        model.ChannelBookmarkLink,
		Emoji:       ":smile:",
	}

	th.App.CreateChannelBookmark(th.Context, bookmark1, "")

	file := &model.FileInfo{
		Id:              model.NewId(),
		CreatorId:       th.BasicUser.Id,
		Path:            "somepath",
		ThumbnailPath:   "thumbpath",
		PreviewPath:     "prevPath",
		Name:            "test file",
		Extension:       "png",
		MimeType:        "images/png",
		Size:            873182,
		Width:           3076,
		Height:          2200,
		HasPreviewImage: true,
	}

	th.App.Srv().Store().FileInfo().Save(file)
	defer th.App.Srv().Store().FileInfo().PermanentDelete(file.Id)

	bookmark2 := &model.ChannelBookmark{
		ChannelId:   th.BasicChannel.Id,
		DisplayName: "Bookmark 2",
		FileId:      file.Id,
		Type:        model.ChannelBookmarkFile,
		Emoji:       ":smile:",
	}

	th.App.CreateChannelBookmark(th.Context, bookmark2, "")

	t.Run("get bookmarks of a channel", func(t *testing.T) {
		bookmarks, err := th.App.GetChannelBookmarks(th.BasicChannel.Id, 0)
		assert.Nil(t, err)
		assert.Len(t, bookmarks, 2)
	})

	t.Run("get bookmarks of a channel after one is deleted (aka only return the changed bookmarks)", func(t *testing.T) {
		now := model.GetMillis()
		th.App.DeleteChannelBookmark(bookmark1.Id, "")

		bookmarks, err := th.App.GetChannelBookmarks(th.BasicChannel.Id, 0)
		assert.Nil(t, err)
		assert.Len(t, bookmarks, 1)

		bookmarks, err = th.App.GetChannelBookmarks(th.BasicChannel.Id, now)
		assert.Nil(t, err)
		assert.Len(t, bookmarks, 1)

		deleted := false
		for _, b := range bookmarks {
			if b.DeleteAt > 0 {
				deleted = true
				break
			}
		}
		assert.Equal(t, deleted, true)
	})
}

func TestUpdateChannelBookmarkSortOrder(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	channelId := th.BasicChannel.Id
	th.Context.Session().UserId = th.BasicUser.Id // set the user for the session

	bookmark0 := &model.ChannelBookmark{
		ChannelId:   channelId,
		DisplayName: "Bookmark 0",
		LinkUrl:     "https://mattermost.com",
		Type:        model.ChannelBookmarkLink,
		Emoji:       ":smile:",
	}

	file := &model.FileInfo{
		Id:              model.NewId(),
		CreatorId:       th.BasicUser.Id,
		Path:            "somepath",
		ThumbnailPath:   "thumbpath",
		PreviewPath:     "prevPath",
		Name:            "test file",
		Extension:       "png",
		MimeType:        "images/png",
		Size:            873182,
		Width:           3076,
		Height:          2200,
		HasPreviewImage: true,
	}

	bookmark1 := &model.ChannelBookmark{
		ChannelId:   channelId,
		DisplayName: "Bookmark 1",
		FileId:      file.Id,
		Type:        model.ChannelBookmarkFile,
		Emoji:       ":smile:",
	}

	_, err := th.App.Srv().Store().FileInfo().Save(file)
	require.NoError(t, err)
	defer th.App.Srv().Store().FileInfo().PermanentDelete(file.Id)

	bookmark2 := &model.ChannelBookmark{
		ChannelId:   channelId,
		DisplayName: "Bookmark 2",
		LinkUrl:     "https://mattermost.com",
		Type:        model.ChannelBookmarkLink,
	}

	bookmark3 := &model.ChannelBookmark{
		ChannelId:   channelId,
		DisplayName: "Bookmark 3",
		LinkUrl:     "https://mattermost.com",
		Type:        model.ChannelBookmarkLink,
	}

	bookmark4 := &model.ChannelBookmark{
		ChannelId:   channelId,
		DisplayName: "Bookmark 4",
		LinkUrl:     "https://mattermost.com",
		Type:        model.ChannelBookmarkLink,
	}

	bookmarkResp, appErr := th.App.CreateChannelBookmark(th.Context, bookmark0, "")
	assert.Nil(t, appErr)
	bookmark0 = bookmarkResp.ChannelBookmark.Clone()

	bookmarkResp, appErr = th.App.CreateChannelBookmark(th.Context, bookmark1, "")
	assert.Nil(t, appErr)
	bookmark1 = bookmarkResp.ChannelBookmark.Clone()

	bookmarkResp, appErr = th.App.CreateChannelBookmark(th.Context, bookmark2, "")
	assert.Nil(t, appErr)
	bookmark2 = bookmarkResp.ChannelBookmark.Clone()

	bookmarkResp, appErr = th.App.CreateChannelBookmark(th.Context, bookmark3, "")
	assert.Nil(t, appErr)
	bookmark3 = bookmarkResp.ChannelBookmark.Clone()

	bookmarkResp, appErr = th.App.CreateChannelBookmark(th.Context, bookmark4, "")
	assert.Nil(t, appErr)
	bookmark4 = bookmarkResp.ChannelBookmark.Clone()

	t.Run("change order of bookmarks first to last", func(t *testing.T) {
		bookmarks, sortErr := th.App.UpdateChannelBookmarkSortOrder(bookmark0.Id, channelId, int64(4), "")
		assert.Nil(t, sortErr)

		assert.Equal(t, find_bookmark(bookmarks, bookmark1.Id).SortOrder, int64(0))
		assert.Equal(t, find_bookmark(bookmarks, bookmark2.Id).SortOrder, int64(1))
		assert.Equal(t, find_bookmark(bookmarks, bookmark3.Id).SortOrder, int64(2))
		assert.Equal(t, find_bookmark(bookmarks, bookmark4.Id).SortOrder, int64(3))
		assert.Equal(t, find_bookmark(bookmarks, bookmark0.Id).SortOrder, int64(4))
	})

	t.Run("change order of bookmarks last to first", func(t *testing.T) {
		bookmarks, sortErr := th.App.UpdateChannelBookmarkSortOrder(bookmark0.Id, channelId, int64(0), "")
		assert.Nil(t, sortErr)

		assert.Equal(t, find_bookmark(bookmarks, bookmark0.Id).SortOrder, int64(0))
		assert.Equal(t, find_bookmark(bookmarks, bookmark1.Id).SortOrder, int64(1))
		assert.Equal(t, find_bookmark(bookmarks, bookmark2.Id).SortOrder, int64(2))
		assert.Equal(t, find_bookmark(bookmarks, bookmark3.Id).SortOrder, int64(3))
		assert.Equal(t, find_bookmark(bookmarks, bookmark4.Id).SortOrder, int64(4))
	})

	t.Run("change order of bookmarks first to third", func(t *testing.T) {
		bookmarks, sortErr := th.App.UpdateChannelBookmarkSortOrder(bookmark0.Id, channelId, int64(2), "")
		assert.Nil(t, sortErr)

		assert.Equal(t, find_bookmark(bookmarks, bookmark1.Id).SortOrder, int64(0))
		assert.Equal(t, find_bookmark(bookmarks, bookmark2.Id).SortOrder, int64(1))
		assert.Equal(t, find_bookmark(bookmarks, bookmark0.Id).SortOrder, int64(2))
		assert.Equal(t, find_bookmark(bookmarks, bookmark3.Id).SortOrder, int64(3))
		assert.Equal(t, find_bookmark(bookmarks, bookmark4.Id).SortOrder, int64(4))

		// now reset order
		th.App.UpdateChannelBookmarkSortOrder(bookmark0.Id, channelId, int64(0), "")
	})

	t.Run("change order of bookmarks second to third", func(t *testing.T) {
		bookmarks, sortErr := th.App.UpdateChannelBookmarkSortOrder(bookmark1.Id, channelId, int64(2), "")
		assert.Nil(t, sortErr)

		assert.Equal(t, find_bookmark(bookmarks, bookmark0.Id).SortOrder, int64(0))
		assert.Equal(t, find_bookmark(bookmarks, bookmark2.Id).SortOrder, int64(1))
		assert.Equal(t, find_bookmark(bookmarks, bookmark1.Id).SortOrder, int64(2))
		assert.Equal(t, find_bookmark(bookmarks, bookmark3.Id).SortOrder, int64(3))
		assert.Equal(t, find_bookmark(bookmarks, bookmark4.Id).SortOrder, int64(4))
	})

	t.Run("change order of bookmarks third to second", func(t *testing.T) {
		bookmarks, sortErr := th.App.UpdateChannelBookmarkSortOrder(bookmark1.Id, channelId, int64(1), "")
		assert.Nil(t, sortErr)

		assert.Equal(t, find_bookmark(bookmarks, bookmark0.Id).SortOrder, int64(0))
		assert.Equal(t, find_bookmark(bookmarks, bookmark1.Id).SortOrder, int64(1))
		assert.Equal(t, find_bookmark(bookmarks, bookmark2.Id).SortOrder, int64(2))
		assert.Equal(t, find_bookmark(bookmarks, bookmark3.Id).SortOrder, int64(3))
		assert.Equal(t, find_bookmark(bookmarks, bookmark4.Id).SortOrder, int64(4))
	})

	t.Run("change order of bookmarks last to previous last", func(t *testing.T) {
		bookmarks, sortErr := th.App.UpdateChannelBookmarkSortOrder(bookmark4.Id, channelId, int64(3), "")
		assert.Nil(t, sortErr)

		assert.Equal(t, find_bookmark(bookmarks, bookmark0.Id).SortOrder, int64(0))
		assert.Equal(t, find_bookmark(bookmarks, bookmark1.Id).SortOrder, int64(1))
		assert.Equal(t, find_bookmark(bookmarks, bookmark2.Id).SortOrder, int64(2))
		assert.Equal(t, find_bookmark(bookmarks, bookmark4.Id).SortOrder, int64(3))
		assert.Equal(t, find_bookmark(bookmarks, bookmark3.Id).SortOrder, int64(4))
	})

	t.Run("change order of bookmarks last to second", func(t *testing.T) {
		bookmarks, sortErr := th.App.UpdateChannelBookmarkSortOrder(bookmark3.Id, channelId, int64(1), "")
		assert.Nil(t, sortErr)

		assert.Equal(t, find_bookmark(bookmarks, bookmark0.Id).SortOrder, int64(0))
		assert.Equal(t, find_bookmark(bookmarks, bookmark3.Id).SortOrder, int64(1))
		assert.Equal(t, find_bookmark(bookmarks, bookmark1.Id).SortOrder, int64(2))
		assert.Equal(t, find_bookmark(bookmarks, bookmark2.Id).SortOrder, int64(3))
		assert.Equal(t, find_bookmark(bookmarks, bookmark4.Id).SortOrder, int64(4))
	})

	t.Run("change order of bookmarks error when new index is out of bounds", func(t *testing.T) {
		_, appErr = th.App.UpdateChannelBookmarkSortOrder(bookmark3.Id, channelId, int64(-1), "")
		assert.Error(t, appErr)
		_, appErr = th.App.UpdateChannelBookmarkSortOrder(bookmark3.Id, channelId, int64(5), "")
		assert.Error(t, appErr)
	})

	t.Run("change order of bookmarks error when bookmark not found", func(t *testing.T) {
		_, appErr = th.App.UpdateChannelBookmarkSortOrder(model.NewId(), channelId, int64(0), "")
		assert.Error(t, appErr)
	})
}