// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import './style.scss';

import React, { useState, useEffect, useRef } from 'react';
import { Popover } from 'antd';
import Tag from 'components/tag';
import { Add, AddRounded } from '@material-ui/icons';
import RemainingTagsList from './remainingTagsList';
import TagsPicker from 'components/tagsPicker';

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

const TagsList = ({ tagsToShow, tags, editable, handleDelete, entityType, entityName, handleTagsUpdate, newEntity = false }) => {
    const [tagsToDisplay, setTagsToDisplay] = useState([]);
    const [remainingTags, setRemainingTags] = useState([]);
    const saveChangesRef = useRef(null);
    const [tagsPop, setTagsPop] = useState(false);

    useEffect(() => {
        if (tags?.length > tagsToShow) {
            const tagsShow = tags?.slice(0, tagsToShow);
            setTagsToDisplay(tagsShow);
            const remainingTagsList = tags?.slice(tagsToShow);
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
                    <div className="plus-tags">
                        <Add className="add" />
                        <p>{remainingTags.length}</p>
                    </div>
                </Popover>
            )}
            {editable && (
                <Popover
                    overlayInnerStyle={tagsPickerPopInnerStyle}
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
                            entity_type={entityType}
                            entity_name={entityName}
                            handleUpdatedTagList={(tags) => {
                                handleTagsUpdate(tags);
                                setTagsPop(false);
                            }}
                            newEntity={newEntity}
                        />
                    }
                >
                    <div className="edit-tags">
                        <AddRounded className="add" />
                        <div className="edit-content">{tags?.length > 0 ? 'Edit tags' : 'Add new tag'}</div>
                    </div>
                </Popover>
            )}
        </div>
    );
};

export default TagsList;
