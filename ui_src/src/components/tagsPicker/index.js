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
import React, { useEffect, useState } from 'react';
import Button from '../../components/button';
import SearchInput from '../searchInput';
import searchIcon from '../../assets/images/searchIcon.svg';
import Modal from '../../components/modal';
import { httpRequest } from '../../services/http';
import { ApiEndpoints } from '../../const/apiEndpoints';
import NewTagGenerator from './newTagGenerator';
import { Add, Check } from '@material-ui/icons';
import { Divider } from 'antd';
import emptyTags from '../../assets/images/emptyTags.svg';

const TagsPicker = ({ tags, entity_id, entity_type, handleUpdatedTagList, handleCloseWithNoChanges }) => {
    const [tagsToDisplay, setTagsToDisplay] = useState([]);
    const [allTags, setAllTags] = useState([]);
    const [searchInput, setSearchInput] = useState('');
    const [checkedList, setCheckedList] = useState(tags);
    const [editedList, setEditedList] = useState(false);
    const [newTagModal, setNewTagModal] = useState(false);

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
            const res = await httpRequest('GET', `${ApiEndpoints.GET_TAGS}`);
            setTagsToDisplay(res);
            setAllTags(res);
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

    const handleSaveChanges = async () => {
        const tagsToRemove = tags.filter((tag) => {
            if (checkedList.some((checkedTag) => tag.name === checkedTag.name)) return false;
            return true;
        });
        var tagsToRemoveNames;
        const tagsToAdd = checkedList.filter((checkedTag) => {
            if (tags.some((tag) => tag.name === checkedTag.name)) return false;
            return true;
        });
        try {
            if (!(tagsToRemove.length === 0)) {
                tagsToRemoveNames = tagsToRemove.map((tag) => {
                    return tag.name;
                });
            }
            const reqBody = {
                tags_to_Add: tagsToAdd,
                tags_to_Remove: tagsToRemoveNames,
                entity_type: entity_type,
                entity_id: entity_id
            };
            const updatedTags = await httpRequest('PUT', `${ApiEndpoints.UPDATE_TAGS_FOR_ENTITY}`, reqBody);
            setEditedList(false);
            handleUpdatedTagList(updatedTags);
        } catch (error) {}
    };

    const handleNewTag = (tag) => {
        setCheckedList([...checkedList, tag]);
        setAllTags([...allTags, tag]);
        setTagsToDisplay([...tagsToDisplay, tag]);
        setEditedList(true);
        setSearchInput('');
        setNewTagModal(false);
    };

    return (
        <div className="tags-picker-wrapper">
            <div className="tags-picker-title">Apply tags to this {entity_type}</div>
            {allTags?.length > 0 && (
                <>
                    <div className="search-input">
                        <SearchInput
                            placeholder="Tag Name"
                            colorType="navy"
                            backgroundColorType="none"
                            borderRadiusType="circle"
                            borderColorType="search-input"
                            iconComponent={<img alt="search tag" src={searchIcon} />}
                            onChange={handleSearch}
                            value={searchInput}
                        />
                    </div>
                    <div className="tags-list">
                        {tagsToDisplay?.length > 0 ? (
                            tagsToDisplay.map((tag) => (
                                <>
                                    <li key={tag.name} className="tag" onClick={() => handleCheck(tag)}>
                                        {checkedList?.some((item) => tag.name === item.name) ? <Check className="checkmark" /> : <div className="no-checkmark"></div>}
                                        <div className="color-circle" style={{ backgroundColor: `rgb(${tag.color})` }}></div>
                                        <div className="tag-name">{tag.name}</div>
                                    </li>
                                    <Divider className="divider" />
                                </>
                            ))
                        ) : (
                            <div className="no-tags">
                                <img className="no-tags-image" alt="empty-tags-list" src={emptyTags} width={80} height={80} />
                                <span className="no-tags-message">No Tags Found</span>
                            </div>
                        )}
                    </div>
                    <div className="create-new-tag" onClick={() => setNewTagModal(true)}>
                        <Add className="add" />
                        <div className="new-button">
                            {`Create New Tag `}
                            {searchInput.length > 0 && `"${searchInput}"`}
                        </div>
                    </div>
                    <Divider className="divider" />
                    <div className="save-cancel-buttons">
                        <Button
                            width={'120px'}
                            height="30px"
                            placeholder={`Cancel`}
                            colorType="black"
                            radiusType="circle"
                            backgroundColorType={'white'}
                            border="gray-light"
                            fontSize="14px"
                            fontWeight="bold"
                            marginRight="5px"
                            onClick={() => {
                                handleCloseWithNoChanges();
                                setSearchInput('');
                            }}
                        />
                        <Button
                            width={'120px'}
                            height="30px"
                            placeholder={`Save`}
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType={'purple'}
                            fontSize="14px"
                            fontWeight="bold"
                            onClick={handleSaveChanges}
                            disabled={!editedList}
                        />
                    </div>
                </>
            )}
            {allTags?.length === 0 && (
                <div className="no-tags">
                    <img className="no-tags-image" alt="empty-tags-list" src={emptyTags} width={80} height={80} />
                    <span className="no-tags-message">No Tags Exist</span>
                    <span className="tags-info-message">Tags will help you organize, search and filter your data</span>
                    <div className="create-new-tag-empty" onClick={() => setNewTagModal(true)}>
                        <Add className="add" />
                        <div className="new-button">{`Create New Tag ${searchInput.length > 0 && tagsToDisplay.length === 0 ? `"${searchInput}"` : ''}`}</div>
                    </div>
                </div>
            )}
            {
                <Modal
                    className="generator-modal"
                    displayButtons={false}
                    height="415px"
                    width="290px"
                    clickOutside={() => setNewTagModal(false)}
                    open={newTagModal}
                    hr={false}
                >
                    <NewTagGenerator searchVal={searchInput} allTags={allTags} handleFinish={(tag) => handleNewTag(tag)} handleCancel={() => setNewTagModal(false)} />
                </Modal>
            }
        </div>
    );
};

export default TagsPicker;
