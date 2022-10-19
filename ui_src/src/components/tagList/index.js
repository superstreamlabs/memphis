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

import React, { useState, useEffect } from 'react';
import { Tag, Dropdown, Space } from 'antd';
import { Add } from '@material-ui/icons';
import Modal from '../modal';
import AllTagsList from './allTagsList';

const TagsList = ({ tags, addNew, handleEdit, closable, handleClose }) => {
    const [tagsToDisplay, setTagsToDisplay] = useState([]);
    const [remainingTags, setRemainingTags] = useState([]);
    const [listTagModal, setListTagModal] = useState(false);
    useEffect(() => {
        if (tags?.length > 3) {
            const tagsShow = tags.slice(0, 3);
            setTagsToDisplay(tagsShow);
            const remainingTagsList = tags.slice(3, tags?.length);
            setRemainingTags(remainingTagsList);
        } else {
            setTagsToDisplay(tags);
        }
    }, []);

    return (
        <div className="tags-list-wrapper">
            {tagsToDisplay?.map((tag, index) => {
                const isLongTag = tag.name.length > 7;
                return (
                    <Tag className="tag-wrapper" key={tag.name} color={tag.color} closable={closable ? closable : false} onClose={() => handleClose(tag.name)}>
                        {isLongTag ? `${tag.name.slice(0, 7)}...` : tag.name}
                    </Tag>
                );
            })}
            {tags?.length > 3 ? (
                <Dropdown
                    overlay={<AllTagsList tags={remainingTags} handleClose={handleClose} closable={closable}></AllTagsList>}
                    trigger={closable ? ['click'] : ['hover']}
                >
                    <Space>
                        <div className="plus-tag-wrapper">
                            <Tag className="tag-wrapper" key={'+'} closable={false} color={'purple'}>
                                +{tags.length - 3}
                            </Tag>
                        </div>
                    </Space>
                </Dropdown>
            ) : (
                <></>
            )}
            {addNew && (
                <div className="edit-tags" onClick={() => handleEdit()}>
                    <Add />
                    <p>Edit Tags</p>
                </div>
            )}
            {/* {
                <Modal
                    header={
                        <div className="audit-header">
                            <p className="title">Tags</p>
                        </div>
                    }
                    displayButtons={false}
                    height="250px"
                    width="140px"
                    clickOutside={() => setListTagModal(false)}
                    open={listTagModal}
                    hr={false}
                >
                    <AllTagsList tags={tags} handleClose={handleClose} closable={closable}></AllTagsList>
                </Modal>
            } */}
        </div>
    );
};

export default TagsList;
