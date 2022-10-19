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
import React, { useEffect, useRef, useState } from 'react';
import Button from '../../components/button';
import { Divider } from 'antd';
import SearchInput from '../searchInput';
import { InfoOutlined } from '@material-ui/icons';
import searchIcon from '../../assets/images/searchIcon.svg';
import Modal from '../../components/modal';
import { httpRequest } from '../../services/http';
import { ApiEndpoints } from '../../const/apiEndpoints';
import NewTagGenerator from './newTagGenerator';
import CheckboxComponent from '../checkBox';

const TagsPicker = ({ tags, handleClick, entity_type, entity_name, handleUpdatedTagList, handleCloseWithNoChanges }) => {
    const [tagsToDisplay, setTagsToDisplay] = useState([]);
    const [allTags, setAllTags] = useState([]);
    const [searchInput, setSearchInput] = useState('');
    const [checkedList, setCheckedList] = useState(tags);
    const [tagsToRemove, setTagsToRemove] = useState([]);
    const [tagsToAdd, setTagsToAdd] = useState([]);
    const [editedList, setEditedList] = useState(false);
    const [newTagModal, setNewTagModal] = useState(false);

    const handleCheck = (e) => {
        const { value, checked } = e.target;
        var updatedList = [...checkedList];
        var tagChecked = allTags.find((tag) => tag.name === value);
        if (checked) {
            updatedList = [...checkedList, tagChecked];
        } else {
            updatedList.splice(checkedList.indexOf(tagChecked), 1);
        }
        setCheckedList(updatedList);
        setEditedList(true);
    };

    useEffect(() => {
        if (tagsToAdd.length > 0 || tagsToRemove.length > 0) {
            setEditedList(true);
        } else {
            setEditedList(false);
        }
    }, [tagsToAdd, tagsToRemove]);

    useEffect(() => {
        const getAllTags = async () => {
            const res = await httpRequest('GET', `${ApiEndpoints.GET_TAGS}`);
            setTagsToDisplay(res);
            setAllTags(res);
        };
        getAllTags();
    }, []);

    useEffect(() => {
        //let allTagsRes = res.filter((allTag) => !tags.some((tag) => allTag.id === tag.id));
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

        const tagsToAdd = checkedList.filter((checkedTag) => {
            if (tags.some((tag) => tag.name === checkedTag.name)) return false;
            return true;
        });
        try {
            if (!(tagsToRemove.length === 0)) {
                var tagsToRemoveNames = tagsToRemove.map((tag) => {
                    return tag.name;
                });
                const reqBody = {
                    names: tagsToRemoveNames,
                    entity_type: entity_type,
                    entity_name: entity_name
                };
                await httpRequest('DELETE', `${ApiEndpoints.REMOVE_TAGS}`, reqBody);
            }
            if (!(tagsToAdd.length === 0)) {
                const reqBody = {
                    tags: tagsToAdd,
                    entity_type: entity_type,
                    entity_name: entity_name
                };
                await httpRequest('POST', `${ApiEndpoints.CREATE_TAGS}`, reqBody);
            }
            setEditedList(false);
            handleUpdatedTagList();
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
            <div className="search-input">
                <SearchInput
                    placeholder="Tag Name"
                    colorType="navy"
                    backgroundColorType="none"
                    width="10vw"
                    height="27px"
                    borderRadiusType="circle"
                    borderColorType="search-input"
                    boxShadowsType="search-input"
                    iconComponent={<img alt="search tag" src={searchIcon} />}
                    onChange={handleSearch}
                    value={searchInput}
                />
            </div>
            <div className="tags-list">
                {tagsToDisplay && tagsToDisplay.length > 0 ? (
                    tagsToDisplay.map((tag) => (
                        <li key={tag.name} className="tag">
                            <input value={tag.name} type="checkbox" defaultChecked={checkedList.some((item) => tag.name === item.name)} onChange={handleCheck} />
                            {/* <CheckboxComponent checkName={tag.name} id={tag.name} checked={checkedList.some((item) => tag.name === item.name)} onChange={handleCheck} /> */}
                            <div className="color-circle" style={{ backgroundColor: tag.color }}></div>
                            <span className="tag-name">{tag.name}</span>
                            <Divider className="divider" />
                        </li>
                    ))
                ) : (
                    <span className="no-new">No Tags With That Name</span>
                )}
            </div>
            <div className="create-new-tag">
                <Button
                    width={'200px'}
                    height="30px"
                    placeholder={`Create New Tag ${searchInput.length > 0 && tagsToDisplay.length === 0 ? searchInput : ''}`}
                    colorType="black"
                    radiusType="circle"
                    backgroundColorType={'none'}
                    fontSize="14px"
                    fontWeight="bold"
                    htmlType="submit"
                    marginLeft="20px"
                    marginBottom="5px"
                    onClick={() => setNewTagModal(true)}
                />
            </div>
            {/* <Button onClick={() => setAuditModal(true)}>Create New Tag {searchInput.length > 0 && tagsToDisplay.length === 0 ? searchInput : ''}</Button> */}
            {/* {editedList && ( */}
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
                    htmlType="submit"
                    marginRight="5px"
                    onClick={handleCloseWithNoChanges}
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
                    htmlType="submit"
                    onClick={handleSaveChanges}
                    disabled={!editedList}
                />

                {/* <Button onClick={handleSaveChanges}>Save Changes</Button> */}
                {/* <Button onClick={handleCloseWithNoChanges}>Cancel</Button> */}
            </div>
            {/* )} */}
            {
                <Modal
                    header={
                        <div className="audit-header">
                            <p className="title">New Tag</p>
                        </div>
                    }
                    displayButtons={false}
                    height="250px"
                    width="320px"
                    clickOutside={() => setNewTagModal(false)}
                    open={newTagModal}
                    hr={false}
                >
                    <NewTagGenerator searchVal={searchInput} allTags={allTags} handleFinish={(tag) => handleNewTag(tag)} />
                </Modal>
            }
        </div>
    );
};

export default TagsPicker;
