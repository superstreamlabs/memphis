// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package server

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
                {content?.username}
            </div>
            <div className="user-type">
                <UserType userType={content?.user_type} />
            </div>
            <div className="user-creation-date">
                <p>{parsingDate(content?.creation_date)} </p>
            </div>
            {content?.user_type !== 'root' && (
                <div className="user-actions">
                    {/* <p>Generate password</p> */}
                    <p onClick={() => modalFlip(true)}>Delete user</p>
                </div>
            )}
            <Modal
                header="Remove user"
                height="120px"
                rBtnText="Cancel"
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
