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
import 'antd/dist/antd.css';
import { PlusOutlined } from '@ant-design/icons';
import { Button, Input, Tag, Tooltip } from 'antd';
import { httpRequest } from '../../services/http';
import { ApiEndpoints } from '../../const/apiEndpoints';

const AddTag = ({ tagList, entity_type, entity_name, handleUpdatedTagList }) => {
    const [tags, setTags] = useState(tagList);
    const [originalTags] = useState(tagList);
    const [allTags, setAllTags] = useState([]);
    const [allTagsToDisplay, setAllTagsToDisplay] = useState([]);
    const [newTags, setNewTags] = useState([]);
    const [deletedTags, setDeletedTags] = useState([]);
    const [inputVisible, setInputVisible] = useState(false);
    const [inputValue, setInputValue] = useState('');
    const [editInputIndex, setEditInputIndex] = useState(-1);
    const [editInputValue, setEditInputValue] = useState('');
    const [editedList, setEditedList] = useState(false);
    const inputRef = useRef(null);
    const editInputRef = useRef(null);

    useEffect(() => {
        if (inputVisible) {
            inputRef.current?.focus();
        }
    }, [inputVisible]);
    useEffect(() => {
        editInputRef.current?.focus();
    }, [inputValue]);

    const handleClose = (removedTag) => {
        debugger;
        console.log(originalTags);
        console.log(tags);
        const newTags = tags.filter((tag) => tag !== removedTag.name);
        setTags(newTags);
        setDeletedTags([...deletedTags, removedTag.name]);
        setEditedList(true);
    };

    const showInput = () => {
        setInputVisible(true);
    };

    const handleInputChange = (e) => {
        setInputValue(e.target.value);
    };

    const handleInputConfirm = () => {
        if (inputValue && tags.indexOf(inputValue) === -1 && !originalTags.some((tag) => tag.name === inputValue) && !tags.some((tag) => tag.name === inputValue)) {
            const newTag = {
                name: inputValue,
                color_bg: 'blue',
                color_txt: 'white'
            };
            setTags([...tags, newTag]);
            setNewTags([...newTags, newTag]);
            setEditedList(true);
        }

        setInputVisible(false);
        setInputValue('');
    };

    const handleEditInputChange = (e) => {
        setEditInputValue(e.target.value);
    };

    const handleEditInputConfirm = () => {
        if (originalTags.some((tag) => tag.name !== editInputValue) && tags.some((tag) => tag.name !== editInputValue)) {
            const newTags = [...tags];
            const newTag = {
                name: editInputValue,
                color_bg: 'blue',
                color_txt: 'white'
            };
            newTags[editInputIndex] = newTag;
            setTags(newTags);
            setEditedList(true);
        }
        setEditInputIndex(-1);
        setInputValue('');
    };

    const handleSaveChanges = async () => {
        try {
            if (!(deletedTags.length === 0)) {
                const reqBody = {
                    names: deletedTags,
                    entity_type: entity_type,
                    entity_name: entity_name
                };
                await httpRequest('DELETE', `${ApiEndpoints.REMOVE_TAGS}`, reqBody);
            }
            if (!(newTags.length === 0)) {
                const reqBody = {
                    tags: newTags,
                    entity_type: entity_type,
                    entity_name: entity_name
                };
                await httpRequest('POST', `${ApiEndpoints.CREATE_TAGS}`, reqBody);
            }
            setDeletedTags([]);
            setNewTags([]);
            setEditedList(false);
            handleUpdatedTagList();
        } catch (error) {}
    };
    return (
        <div className="add-tag-container">
            <div className="existing-tags">
                {tags.map((tag, index) => {
                    if (editInputIndex === index) {
                        return (
                            <Input
                                ref={editInputRef}
                                key={tag.name}
                                size="small"
                                className="tag-input"
                                value={editInputValue}
                                onChange={handleEditInputChange}
                                onBlur={handleEditInputConfirm}
                                onPressEnter={handleEditInputConfirm}
                            />
                        );
                    }
                    const isLongTag = tag.name.length > 5;
                    const tagElem = (
                        <Tag className="edit-tag" key={tag.name} color={tag.color_bg} closable={true} onClose={() => handleClose(tag)}>
                            <span
                            // onDoubleClick={(e) => {
                            //     if (index !== 0) {
                            //         setEditInputIndex(index);
                            //         setEditInputValue(tag.name);
                            //         e.preventDefault();
                            //     }
                            // }}
                            >
                                {isLongTag ? `${tag.name.slice(0, 5)}...` : tag.name}
                            </span>
                        </Tag>
                    );
                    return isLongTag ? (
                        <Tooltip title={tag.name} key={tag.name}>
                            {tagElem}
                        </Tooltip>
                    ) : (
                        tagElem
                    );
                })}
            </div>
            {inputVisible && (
                <Input
                    ref={inputRef}
                    type="text"
                    size="small"
                    className="tag-input"
                    value={inputValue}
                    onChange={handleInputChange}
                    onBlur={handleInputConfirm}
                    onPressEnter={handleInputConfirm}
                />
            )}
            {!inputVisible && (
                <Tag className="site-tag-plus" onClick={showInput}>
                    <PlusOutlined /> New Tag
                </Tag>
            )}
            {
                <div className="tags-list">
                    {allTagsToDisplay.map((tag, index) => {
                        const isLongTag = tag.name.length > 5;
                        const tagElem = (
                            <Tag className="edit-tag" key={tag.name} color={tag.color_bg} closable={false} onClick={() => handleClose(tag)}>
                                <span>{isLongTag ? `${tag.name.slice(0, 5)}...` : tag.name}</span>
                            </Tag>
                        );
                        return isLongTag ? (
                            <Tooltip title={tag.name} key={tag.name}>
                                {tagElem}
                            </Tooltip>
                        ) : (
                            tagElem
                        );
                    })}
                </div>
            }
            {editedList && <Button onClick={handleSaveChanges}>Save Changes</Button>}
        </div>
    );
};

export default AddTag;
