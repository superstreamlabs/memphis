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

import React, { useState, useEffect, useRef } from 'react';
import { Space, Popover } from 'antd';
import Tag from '../tag';
import { Add, AddRounded } from '@material-ui/icons';
import RemainingTagsList from './remainingTagsList';
import TagsPicker from '../tagsPicker';

const remainingTagsPopInnerStyle = { maxWidth: '155px', padding: '10px', paddingBottom: '10px', borderRadius: '12px', border: '1px solid #f0f0f0' };
const tagsPickerPopInnerStyle = {
    width: '250px',
    height: '313px',
    borderRadius: '8px',
    border: '1px solid #E4E4E4',
    padding: '0px 0px',
    overflow: 'hidden',
    boxShadow: '0px 23px 44px rgba(176, 183, 195, 0.14)'
};

const TagsList = ({ tagsToShow, tags, editable, handleDelete, entityName, entityID, handleTagsUpdate, newEntity = false }) => {
    const [tagsToDisplay, setTagsToDisplay] = useState([]);
    const [remainingTags, setRemainingTags] = useState([]);
    const saveChangesRef = useRef(null);
    const [tagsPop, setTagsPop] = useState(false);

    useEffect(() => {
        if (tags?.length > tagsToShow) {
            const tagsShow = tags.slice(0, tagsToShow);
            setTagsToDisplay(tagsShow);
            const remainingTagsList = tags.slice(tagsToShow);
            setRemainingTags(remainingTagsList);
        } else {
            setTagsToDisplay(tags);
            setRemainingTags([]);
        }
    }, [tags, tagsToShow]);

    const handleOpenChange = (newOpen) => {
        if (!newOpen) saveChangesRef?.current.handleSaveChanges();
        setTagsPop(newOpen);
    };

    return (
        <div className="tags-list-wrapper">
            {tagsToDisplay?.map((tag, index) => {
                return <Tag key={index} tag={tag} editable={editable || false} onDelete={() => handleDelete(tag.name)} />;
            })}
            {remainingTags?.length > 0 && (
                <Popover
                    overlayInnerStyle={remainingTagsPopInnerStyle}
                    placement="bottomLeft"
                    content={<RemainingTagsList tags={remainingTags} handleDelete={(tag) => handleDelete(tag)} editable={editable}></RemainingTagsList>}
                >
                    <Space className="space">
                        <div className="plus-tags">
                            <Add className="add" />
                            <p>{remainingTags.length}</p>
                        </div>
                    </Space>
                </Popover>
            )}
            {editable && (
                <Popover
                    overlayInnerStyle={tagsPickerPopInnerStyle}
                    zIndex={2}
                    destroyTooltipOnHide={true}
                    trigger="click"
                    placement="bottomLeft"
                    open={tagsPop}
                    onOpenChange={(open) => {
                        handleOpenChange(open);
                    }}
                    content={
                        <TagsPicker
                            ref={saveChangesRef}
                            tags={tags}
                            entity_id={entityID}
                            entity_type={'station'}
                            handleUpdatedTagList={(tags) => {
                                handleTagsUpdate(tags);
                                setTagsPop(false);
                            }}
                            entityName={entityName}
                            newEntity={newEntity}
                        />
                    }
                >
                    <Space className="space">
                        <div className="edit-tags">
                            <AddRounded className="add" />
                            <div className="edit-content">{tags?.length > 0 ? 'Edit tags' : 'Add new tag'}</div>
                        </div>
                    </Space>
                </Popover>
            )}
        </div>
    );
};

export default TagsList;
