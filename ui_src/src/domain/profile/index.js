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

import React, { useEffect, useContext, useState } from 'react';

import { LOCAL_STORAGE_AVATAR_ID, USER_IMAGE } from '../../const/localStorageConsts';
import { ApiEndpoints } from '../../const/apiEndpoints';
import { httpRequest } from '../../services/http';
import { Context } from '../../hooks/store';

function Profile() {
    const [state, dispatch] = useContext(Context);
    const [avatar, setAvatar] = useState(1);
    const [imageUrl, setImageUrl] = useState(localStorage.getItem(USER_IMAGE));

    useEffect(() => {
        const storedImageUrl = localStorage.getItem(USER_IMAGE);
        setImageUrl(storedImageUrl);
    }, []);

    useEffect(() => {
        dispatch({ type: 'SET_ROUTE', payload: 'profile' });
        setAvatar(Number(localStorage.getItem(LOCAL_STORAGE_AVATAR_ID)) || state?.userData?.avatar_id);
    }, []);

    const editAvatar = async (avatarId) => {
        try {
            const data = await httpRequest('PUT', `${ApiEndpoints.EDIT_AVATAR}`, { avatar_id: avatarId });
            setAvatar(data.avatar_id);
            localStorage.setItem(LOCAL_STORAGE_AVATAR_ID, data.avatar_id);
            dispatch({ type: 'SET_AVATAR_ID', payload: data.avatar_id });
        } catch (err) {
            return;
        }
    };

    return (
        <div className="profile-page">
            <div className="profile-container">
                <div className="header-preferences">
                    <p className="main-header">Edit profile</p>
                    <p className="memphis-label">Modify your profile information and preferences</p>
                </div>
                <div className="avatar-section">
                    <p className="title">Avatar</p>
                    <div className="avatar-images">
                        {localStorage.getItem(USER_IMAGE) && localStorage.getItem(USER_IMAGE) !== 'undefined' && (
                            <div className={'avatar-img selected'}>
                                <img className="avatar-image" src={imageUrl} width={35} height={35} alt="avater" />
                            </div>
                        )}
                        {Array.from(Array(8).keys()).map((item, index) => {
                            return (
                                <div
                                    key={index}
                                    className={
                                        localStorage.getItem(USER_IMAGE) && localStorage.getItem(USER_IMAGE) !== 'undefined'
                                            ? 'avatar-img avatar-disable'
                                            : avatar === item + 1
                                            ? 'avatar-img selected'
                                            : 'avatar-img'
                                    }
                                    onClick={() => (localStorage.getItem(USER_IMAGE) === 'undefined' || !localStorage.getItem(USER_IMAGE)) && editAvatar(item + 1)}
                                >
                                    <img src={require(`../../assets/images/bots/avatar${item + 1}.svg`)} alt="avater" />
                                </div>
                            );
                        })}
                    </div>
                </div>
            </div>
        </div>
    );
}

export default Profile;
