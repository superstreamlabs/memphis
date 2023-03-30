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

import React, { useEffect, useState } from 'react';
import UserType from './userType';
import { httpRequest } from '../../../services/http';
import { ApiEndpoints } from '../../../const/apiEndpoints';
import Modal from '../../../components/modal';
import { parsingDate } from '../../../services/valueConvertor';

function UserItem({ content, handleRemoveUser }) {
    const defaultAvatarId = 1;
    const [avatarUrl, setAvatarUrl] = useState(1);
    const [open, modalFlip] = useState(false);

    useEffect(() => {
        setAvatarImage(content?.avatar_id || defaultAvatarId);
    }, [content]);

    const setAvatarImage = (avatarId) => {
        setAvatarUrl(require(`../../../assets/images/bots/avatar${avatarId}.svg`));
    };

    const removeUser = async (username) => {
        try {
            await httpRequest('DELETE', ApiEndpoints.REMOVE_USER, {
                username: username
            });
            handleRemoveUser();
        } catch (error) {}
    };
    return (
        <div className="users-item">
            <div className="user-name">
                <div className="user-avatar">
                    <img src={avatarUrl} width={25} height={25} alt="avatar" />
                </div>
                <p>{content?.username}</p>
            </div>
            <div className="user-type">
                <UserType userType={content?.user_type} />
            </div>
            <div className="user-creation-date">
                <p>{parsingDate(content?.created_at)} </p>
            </div>
            {content?.user_type !== 'root' && (
                <div className="user-actions">
                    {/* <p>Generate password</p> */}
                    <p onClick={() => modalFlip(true)}>Delete user</p>
                </div>
            )}
            <Modal
                header="Delete user"
                height="120px"
                rBtnText="Delete"
                lBtnText="Remove"
                lBtnClick={() => {
                    removeUser(content?.username);
                }}
                clickOutside={() => modalFlip(false)}
                rBtnClick={() => modalFlip(false)}
                open={open}
            >
                <label>
                    Are you sure you want to delete "<b>{content?.username}</b>"?
                </label>
                <br />
            </Modal>
        </div>
    );
}
export default UserItem;
