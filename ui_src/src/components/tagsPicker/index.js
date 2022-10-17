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
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import './style.scss';
import React, { useEffect, useRef, useState } from 'react';
import Button from '../../components/button';
import { Badge, Divider, Space } from 'antd';
import SearchInput from '../searchInput';
import { InfoOutlined } from '@material-ui/icons';
import searchIcon from '../../assets/images/searchIcon.svg';
import Modal from '../../components/modal';
import { httpRequest } from '../../services/http';
import { ApiEndpoints } from '../../const/apiEndpoints';
import NewTagGenerator from './newTagGenerator';

const TagsPicker = ({ tags, handleClick, entity_type, entity_name, handleUpdatedTagList, handleCloseWithNoChanges }) => {
    const [tagsToDisplay, setTagsToDisplay] = useState([]);
    const [allTags, setAllTags] = useState([]);
    const [selectedTags, setSelectedTags] = useState(tags);
    const [searchInput, setSearchInput] = useState('');
    const [checked, setChecked] = useState(tags);
    const [tagsToRemove, setTagsToRemove] = useState([]);
    const [tagsToAdd, setTagsToAdd] = useState([]);
    const [newTags, setNewTags] = useState([]);
    const [editedList, setEditedList] = useState(false);
    const [newTagModal, setNewTagModal] = useState(false);

    const handleCheck = (event) => {
        var updatedList = [...checked];
        var updatedTagsToRemove = [...tagsToRemove];
        if (event.target.checked) {
            updatedList = [...checked, event.target.value];
            if (selectedTags.some((tag) => event.target.value === tag.name)) {
                updatedTagsToRemove.splice(tagsToRemove.indexOf(event.target.value), 1);
                setTagsToRemove(updatedTagsToRemove);
            } else {
                let tagToAdd = allTags.find((tag) => {
                    return tag.name === event.target.value;
                });
                if (tagToAdd) {
                    setTagsToAdd([...tagsToAdd, tagToAdd]);
                }
            }
        } else {
            updatedList.splice(checked.indexOf(event.target.value), 1);
            if (selectedTags.some((item) => event.target.value === item.name)) {
                updatedTagsToRemove = [...tagsToRemove, event.target.value];
                setTagsToRemove(updatedTagsToRemove);
            }
        }
        setChecked(updatedList);
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
            // let allTagsRes = res.filter((allTag) => !tags.some((tag) => allTag.id === tag.id));
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

    // useEffect(() => {
    //     // setTagsToDisplay([...allTags, newTags]);
    //     setChecked([...checked, newTags]);
    // }, [newTags]);

    const handleSearch = (e) => {
        setSearchInput(e.target.value);
    };

    const handleSaveChanges = async () => {
        try {
            if (!(tagsToRemove.length === 0)) {
                const reqBody = {
                    names: tagsToRemove,
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
            setTagsToRemove([]);
            setTagsToAdd([]);
            setNewTags([]);
            setEditedList(false);
            handleUpdatedTagList();
        } catch (error) {}
    };

    const handleNewTag = (tag) => {
        setChecked([...checked, newTags]);
        setNewTagModal(false);
        setSearchInput('');
        setTagsToAdd([...tagsToAdd, tag]);
        setNewTags([...newTags, tag]);
        setTagsToDisplay([...allTags, tag]);
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
                        <li key={tag.id} className="tag">
                            <input value={tag.name} type="checkbox" defaultChecked={checked.some((item) => tag.name === item.name)} onChange={handleCheck} />
                            <div className="color-circle" style={{ backgroundColor: tag.color }}></div>
                            <span className="tag-name">{tag.name}</span>
                            <hr></hr>
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
            {editedList && (
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
                    />

                    {/* <Button onClick={handleSaveChanges}>Save Changes</Button> */}
                    {/* <Button onClick={handleCloseWithNoChanges}>Cancel</Button> */}
                </div>
            )}
            {
                <Modal
                    header={
                        <div className="audit-header">
                            <p className="title">New Tag</p>
                        </div>
                    }
                    displayButtons={false}
                    height="300px"
                    width="300px"
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
