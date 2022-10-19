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
import TagsList from '../tagList';
import TagsPicker from '../tagsPicker';

const EditTags = ({ tagList, entity_type, entity_name, handleUpdatedTagList }) => {
    const [tags, setTags] = useState(tagList);
    const [originalTags] = useState(tagList);
    const [allTags, setAllTags] = useState([]);
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

    useEffect(() => {
        const getAllTags = async () => {
            const res = await httpRequest('GET', `${ApiEndpoints.GET_TAGS}`);
            // let allTagsRes = res.filter((allTag) => !tags.some((tag) => allTag.id === tag.id));
            setAllTags(res);
        };
        getAllTags();
    }, []);

    const handleClose = (removedTag) => {
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
                color: 'blue'
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
                color: 'blue'
            };
            newTags[editInputIndex] = newTag;
            setTags(newTags);
            setEditedList(true);
        }
        setEditInputIndex(-1);
        setInputValue('');
    };

    const handleTagClick = (tag) => {
        setTags(...tags, tag);
        let allTagsEdit = allTags;
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
            {
                <div className="tags-list">
                    <TagsPicker tags={tags} />
                </div>
            }
            {editedList && (
                <div>
                    <div>
                        <Button onClick={handleSaveChanges}>Save Changes123</Button>
                    </div>
                    <div>
                        <Button>Cancel</Button>
                    </div>
                </div>
            )}
        </div>
    );
};

export default EditTags;
