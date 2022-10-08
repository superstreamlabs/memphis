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

import React, { useEffect, useContext, useState } from 'react';
import { Context } from '../../../hooks/store';
import pathDomains from '../../../router';
import { LOCAL_STORAGE_ALLOW_ANALYTICS, LOCAL_STORAGE_USER_NAME, LOCAL_STORAGE_COMPANY_LOGO, LOCAL_STORAGE_AVATAR_ID } from '../../../const/localStorageConsts';
import { Checkbox, Divider, Upload, message } from 'antd';
import RadioButton from '../../../components/radioButton';
import Button from '../../../components/button';
import Modal from '../../../components/modal';
import { ApiEndpoints } from '../../../const/apiEndpoints';
import { httpRequest } from '../../../services/http';
import Logo from '../../../assets/images/logo.svg';

function Profile() {
    const [userName, setUserName] = useState('');
    const [state, dispatch] = useContext(Context);
    const [avatar, setAvatar] = useState(1);
    const [open, modalFlip] = useState(false);
    const [allowAnalytics, setAllowAnalytics] = useState();
    const [checkboxdeleteAccount, setCheckboxdeleteAccount] = useState(false);

    function getBase64(file) {
        return new Promise((resolve, reject) => {
            const reader = new FileReader();
            reader.readAsDataURL(file);
            reader.onload = () => resolve(reader.result);
            reader.onerror = (error) => reject(error);
        });
    }

    const [loading, setLoading] = useState(false);
    const [previewVisible, setpreviewVisible] = useState(false);
    const [previewTitle, setpreviewTitle] = useState('');
    const [previewImage, setpreviewImage] = useState('');
    const [fileList, updateFileList] = useState(
        localStorage.getItem(LOCAL_STORAGE_COMPANY_LOGO)
            ? [
                  {
                      uid: '1',
                      name: 'company_logo',
                      status: 'done',
                      url: localStorage.getItem(LOCAL_STORAGE_COMPANY_LOGO)
                  }
              ]
            : []
    );

    const uploadLogo = async ({ file, onSuccess, onError }) => {
        let dataImg = new FormData();
        dataImg.append('file', file);
        try {
            const data = await httpRequest('PUT', ApiEndpoints.EDIT_COMPANY_LOGO, dataImg);
            localStorage.setItem(LOCAL_STORAGE_COMPANY_LOGO, data.image);
            dispatch({ type: 'SET_COMPANY_LOGO', payload: data.image });
            onSuccess('ok');
        } catch (err) {
            onError('error');
        }
    };
    const deleteLogo = async ({ onSuccess, onError }) => {
        try {
            await httpRequest('DELETE', ApiEndpoints.REMOVE_COMPANY_LOGO);
            localStorage.setItem(LOCAL_STORAGE_COMPANY_LOGO, '');
            dispatch({ type: 'SET_COMPANY_LOGO', payload: '' });
            onSuccess('ok');
        } catch (err) {
            onError('error');
        }
    };
    const handleCancel = () => {
        setpreviewVisible(false);
    };

    const uploadProps = {
        beforeUpload: (file) => {
            const isJpgOrPng = file.type === 'image/jpeg' || file.type === 'image/png';
            if (!isJpgOrPng) {
                message.error('JPG/PNG format required', 3);
            }
            return isJpgOrPng;
        },
        customRequest: (file) => uploadLogo(file),
        onChange(info) {
            let fileList = info.fileList.filter((file) => !!file.status);
            fileList = fileList.slice(-1);
            updateFileList(fileList);
            handleMessage(info.file);
        },
        onRemove: (file) => {
            deleteLogo(file);
        }
    };

    const handleMessage = (file) => {
        if (file.response === 'ok' && file.status === 'done') {
            message.success(`Done uploading ${file.name} file`, 3);
        } else if (file.response === 'error' && file.status === 'done') {
            message.error(`${file.name} file upload failed`, 3);
        }
        if (file.response === 'ok' && file.status === 'removed') {
            message.success(`Removed uploading ${file.name} file`, 3);
        } else if (file.response === 'error' && file.status === 'removed') {
            message.error(`${file.name} file upload failed`, 3);
        }
    };

    const onPreview = async (file) => {
        if (!file.url && !file.preview) {
            file.preview = await getBase64(file.originFileObj);
        }
        setpreviewVisible(true);
        setpreviewImage(file.url || file.preview);
        setpreviewTitle(file.name || file.url.substring(file.url.lastIndexOf('/') + 1));
    };
    useEffect(() => {
        setUserName(localStorage.getItem(LOCAL_STORAGE_USER_NAME));
        setAvatar(localStorage.getItem('profile_pic') || state?.userData?.avatar_id || Number(localStorage.getItem(LOCAL_STORAGE_AVATAR_ID))); // profile_pic is available only in sandbox env
        setAllowAnalytics(localStorage.getItem(LOCAL_STORAGE_ALLOW_ANALYTICS) === 'false' ? false : true);
    }, []);

    const removeMyUser = async () => {
        try {
            await httpRequest('DELETE', `${ApiEndpoints.REMOVE_MY_UER}`);
            modalFlip(false);
            localStorage.clear();
            window.location.assign(pathDomains.login);
        } catch (err) {
            return;
        }
    };

    const sendAnalytics = async (analyticsFlag) => {
        try {
            await httpRequest('PUT', `${ApiEndpoints.EDIT_ANALYTICS}`, { send_analytics: analyticsFlag });
            setAllowAnalytics(analyticsFlag);
            localStorage.setItem(LOCAL_STORAGE_ALLOW_ANALYTICS, analyticsFlag);
        } catch (err) {
            return;
        }
    };

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
        <div className="profile-container">
            <div className="header">
                <p className="main-header">Profile</p>
                <p className="sub-header">Select your avater for showing in this system</p>
            </div>
            <div className="avatar-section">
                <p className="title">Avatar</p>
                <div className="avatar-images">
                    {Array.from(Array(9).keys()).map((item) => {
                        return (
                            <div
                                className={process.env.REACT_APP_SANDBOX_ENV ? 'sub-icon-wrapper-sandbox' : avatar === item + 1 ? 'avatar-img selected' : 'avatar-img'}
                                onClick={process.env.REACT_APP_SANDBOX_ENV ? '' : () => editAvatar(item + 1)}
                            >
                                <img
                                    src={localStorage.getItem('profile_pic') || require(`../../../assets/images/bots/avatar${item + 1}.svg`)} // profile_pic is available only in sandbox env
                                    width={localStorage.getItem('profile_pic') ? 35 : ''}
                                    height={localStorage.getItem('profile_pic') ? 35 : ''}
                                    alt="avater"
                                />
                            </div>
                        );
                    })}
                </div>
            </div>
            <div className="company-logo-section">
                <p className="title">Company Logo</p>
                <div className="company-logo">
                    <img className="logoimg" src={state?.companyLogo || Logo} alt="companyLogo" />
                    <div className="company-logo-right">
                        <div className="update-remove-logo">
                            <Upload name="avatar" {...uploadProps} showUploadList={false} onPreview={onPreview} fileList={fileList}>
                                <Button
                                    className="modal-btn"
                                    width="160px"
                                    height="36px"
                                    placeholder="Upload New"
                                    colorType="white"
                                    radiusType="circle"
                                    backgroundColorType="purple"
                                    fontSize="14px"
                                    fontWeight="600"
                                    aria-haspopup="true"
                                />
                            </Upload>
                            <Button
                                className="modal-btn"
                                width="200px"
                                height="36px"
                                placeholder="Remove Current Logo"
                                colorType="red"
                                radiusType="circle"
                                backgroundColorType="none"
                                border="gray"
                                boxShadowsType="gray"
                                fontSize="14px"
                                fontWeight="600"
                                aria-haspopup="true"
                                onClick={() => deleteLogo(fileList[0])}
                            />
                        </div>
                        <label className="company-logo-description">Logo must be 200x200 pixel and size is less than 5mb</label>
                    </div>
                </div>
            </div>
            <Divider />
            <div className="analytics-section">
                <p className="title">Analytics</p>
                <label className="analytics-description">Lorem Ipsum is simply dummy text of the printing and typesetting industry.</label>
                <div className="radioButton-section">
                    <RadioButton
                        options={[
                            { id: 0, value: true, label: 'Allow Analytics' },
                            { id: 1, value: false, label: 'Don’t allow any analytics' }
                        ]}
                        radioValue={allowAnalytics}
                        onChange={(e) => sendAnalytics(e.target.value)}
                        labelType
                    />
                </div>
            </div>
            <Divider />
            <div className="delete-account-section">
                <p className="title">Delete your account</p>
                <label className="delete-account-description">
                    When you delete your account, you lose access to Front account services, and we permanently delete your personal data. You can cancel the deletion for
                    14 days.
                </label>
                <div className="delete-account-checkbox">
                    <Checkbox checked={checkboxdeleteAccount} onChange={() => setCheckboxdeleteAccount(!checkboxdeleteAccount)} name="delete-account" />
                    <p onClick={() => setCheckboxdeleteAccount(!checkboxdeleteAccount)}>Confirm that I want to delete my account.</p>
                </div>
                <Button
                    className="modal-btn"
                    width="200px"
                    height="36px"
                    placeholder="Delete Account"
                    colorType="white"
                    radiusType="circle"
                    backgroundColorType="red"
                    border="red"
                    boxShadowsType="red"
                    fontSize="14px"
                    fontWeight="600"
                    aria-haspopup="true"
                    disabled={!checkboxdeleteAccount || userName === 'root'}
                    onClick={() => modalFlip(true)}
                />
            </div>
            <Modal
                header="Remove accont"
                height="120px"
                rBtnText="Cancel"
                lBtnText="Remove"
                lBtnClick={() => {
                    removeMyUser();
                }}
                clickOutside={() => modalFlip(false)}
                rBtnClick={() => modalFlip(false)}
                open={open}
            >
                <label>Are you sure you want to remove user account?</label>
                <br />
                <label>Please note that this action is irreversible.</label>
            </Modal>
        </div>
    );
}

export default Profile;
