// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO e SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import './style.scss';
import React, { forwardRef, useEffect, useImperativeHandle, useState } from 'react';
import SearchInput from '../searchInput';
import { ReactComponent as SearchIcon } from '../../assets/images/searchIcon.svg';
import Modal from '../../components/modal';
import { httpRequest } from '../../services/http';
import { ApiEndpoints } from '../../const/apiEndpoints';
import NewTagGenerator from './newTagGenerator';
import { AddRounded, Check } from '@material-ui/icons';
import { Divider } from 'antd';
import { ReactComponent as EmptyTagsIcon } from '../../assets/images/emptyTags.svg';
import Loader from '../loader';

const TagsPicker = forwardRef(({ tags, entity_name, entity_type, handleUpdatedTagList, newEntity = false }, ref) => {
    const [tagsToDisplay, setTagsToDisplay] = useState([]);
    const [allTags, setAllTags] = useState([]);
    const [searchInput, setSearchInput] = useState('');
    const [checkedList, setCheckedList] = useState(tags);
    const [editedList, setEditedList] = useState(false);
    const [newTagModal, setNewTagModal] = useState(false);
    const [getTagsLoading, setGetTagsLoading] = useState(true);

    const handleCheck = (tagToHandle) => {
        const checked = checkedList?.some((item) => tagToHandle.name === item.name);
        if (checked) {
            setCheckedList(checkedList.filter((item) => item.name !== tagToHandle.name));
        } else {
            setCheckedList([...checkedList, tagToHandle]);
        }
        setEditedList(true);
    };

    useEffect(() => {
        const getAllTags = async () => {
            try {
                const res = await httpRequest('GET', `${ApiEndpoints.GET_TAGS}`);
                setTagsToDisplay(res);
                setAllTags(res);
                setGetTagsLoading(false);
            } catch (error) {
                setGetTagsLoading(false);
            }
        };
        getAllTags();
    }, []);

    useEffect(() => {
        if (searchInput.length > 0) {
            const results = allTags.filter((tag) => {
                return tag.name.toLowerCase().startsWith(searchInput.toLowerCase());
            });
            setTagsToDisplay(results);
        } else {
            setTagsToDisplay(allTags);
        }
    }, [searchInput]);

    const handleSearch = (e) => {
        setSearchInput(e.target.value);
    };

    useImperativeHandle(ref, () => ({
        async handleSaveChanges() {
            if (editedList) {
                const tagsToAdd = checkedList.filter((checkedTag) => {
                    if (tags.some((tag) => tag.name === checkedTag.name)) return false;
                    return true;
                });
                if (newEntity) {
                    handleUpdatedTagList(checkedList);
                } else {
                    const tagsToRemove = tags.filter((tag) => {
                        if (checkedList.some((checkedTag) => tag.name === checkedTag.name)) return false;
                        return true;
                    });
                    var tagsToRemoveNames;
                    try {
                        if (!(tagsToRemove.length === 0)) {
                            tagsToRemoveNames = tagsToRemove.map((tag) => {
                                return tag.name;
                            });
                        }
                        const reqBody = {
                            tags_to_add: tagsToAdd,
                            tags_to_remove: tagsToRemoveNames,
                            entity_type: entity_type,
                            entity_name: entity_name
                        };
                        const updatedTags = await httpRequest('PUT', `${ApiEndpoints.UPDATE_TAGS_FOR_ENTITY}`, reqBody);
                        setEditedList(false);
                        setSearchInput('');
                        handleUpdatedTagList(updatedTags);
                    } catch (error) {}
                }
            }
        }
    }));

    const handleNewTag = (tag) => {
        setCheckedList([tag, ...checkedList]);
        setAllTags([tag, ...allTags]);
        setTagsToDisplay([tag, ...tagsToDisplay]);
        setEditedList(true);
        setSearchInput('');
        setNewTagModal(false);
    };

    return (
        <div className="tags-picker-wrapper">
            <div className="tags-picker-title">Apply tags</div>
            {getTagsLoading && (
                <div className="loader-uploading">
                    <Loader />
                </div>
            )}
            {!getTagsLoading && allTags?.length > 0 && (
                <>
                    <div className="search-input">
                        <SearchInput
                            placeholder="Tag name"
                            colorType="navy"
                            backgroundColorType="none"
                            borderRadiusType="circle"
                            borderColorType="search-input"
                            iconComponent={<SearchIcon alt="search tag" />}
                            onChange={handleSearch}
                            value={searchInput}
                        />
                    </div>
                    <div className="tags-list">
                        {tagsToDisplay?.length > 0 ? (
                            tagsToDisplay.map((tag, index) => (
                                <div key={index}>
                                    <li key={index} className="tag" onClick={() => handleCheck(tag)}>
                                        {<Check className="checkmark" style={!checkedList?.some((item) => tag.name === item.name) ? { color: 'transparent' } : {}} />}
                                        <div className="color-circle" style={{ backgroundColor: `rgb(${tag.color})` }}></div>
                                        <div className="tag-name">{tag.name}</div>
                                    </li>
                                    <Divider className="divider" />
                                </div>
                            ))
                        ) : (
                            <div className="no-tags">
                                <EmptyTagsIcon className="no-tags-image" alt="empty-tags-list" width={80} height={80} />
                                <span className="no-tags-message">No tags found</span>
                            </div>
                        )}
                    </div>
                    {tagsToDisplay?.length < 5 && <Divider className="divider" />}
                    <div className="create-new-tag" onClick={() => setNewTagModal(true)}>
                        <AddRounded className="add" />
                        <p className="new-button">
                            Create new tag {` `}
                            {searchInput.length > 0 && `"`}
                            <span className="create-new-search">{searchInput.length > 0 && `${searchInput}`}</span>
                            {searchInput.length > 0 && `"`}
                        </p>
                    </div>
                </>
            )}
            {!getTagsLoading && allTags?.length === 0 && (
                <div className="no-tags">
                    <EmptyTagsIcon className="no-tags-image" alt="empty-tags-list" width={80} height={80} />
                    <span className="no-tags-message">No tags exist</span>
                    <span className="tags-info-message">Tags will help you control, group, search, and filter your different entities</span>
                    <div className="create-new-tag-empty" onClick={() => setNewTagModal(true)}>
                        <AddRounded className="add" />
                        <div className="new-button">{`Create new tag`}</div>
                    </div>
                </div>
            )}
            {
                <Modal
                    className="generator-modal"
                    displayButtons={false}
                    width="252px"
                    clickOutside={() => setNewTagModal(false)}
                    open={newTagModal}
                    hr={false}
                    zIndex="9999"
                >
                    <NewTagGenerator searchVal={searchInput} allTags={allTags} handleFinish={(tag) => handleNewTag(tag)} handleCancel={() => setNewTagModal(false)} />
                </Modal>
            }
        </div>
    );
});

export default TagsPicker;
