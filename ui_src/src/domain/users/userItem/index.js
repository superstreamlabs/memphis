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
                    <img src={avatarUrl} width={25} height={25} alt="avatar"></img>
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
