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

import React, { useContext, useState } from 'react';
import { Upload, message, Modal } from 'antd';
import { LoadingOutlined } from '@ant-design/icons';

import { httpRequest } from '../../../../services/http';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import { LOCAL_STORAGE_COMPANY_LOGO } from '../../../../const/localStorageConsts';
import { Context } from '../../../../hooks/store';

function getBase64(file) {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.readAsDataURL(file);
        reader.onload = () => resolve(reader.result);
        reader.onerror = (error) => reject(error);
    });
}

const ImgLoader = () => {
    const [state, dispatch] = useContext(Context);
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

    const uploadButton = (
        <div className="dad-sentence">
            {loading && <LoadingOutlined />}
            <div className="gray-line">Drag &#38; drop to upload </div>
            <div className="purple-line">or browse </div>
        </div>
    );
    return (
        <div>
            <Upload name="avatar" listType="picture-card" className="avatar-uploader" {...uploadProps} onPreview={onPreview} fileList={fileList}>
                {fileList.length > 0 ? null : uploadButton}
            </Upload>
            <Modal visible={previewVisible} title={previewTitle} footer={null} onCancel={handleCancel} clickOutside={handleCancel}>
                <img alt="example" style={{ width: '100%' }} src={previewImage} />
            </Modal>
        </div>
    );
};

export default ImgLoader;
